package console

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/elgris/sqrl"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
	"github.com/lokhman/yams/yams"
	"gopkg.in/go-playground/validator.v9"
)

const (
	errCodeUnknown = (iota + 0xFF) & 0xFF
	_              // errCodeUndefined
	errCodeAuthNoHeader
	errCodeAuthBadToken
	errCodeAuthNoUser
	errCodeAuthFailedACL
	errCodeBadCredentials
	errCodeInvalidIdentifier
	errCodeUsernameExists
	errCodeProfileHostExists
	errCodeInvalidAdapter
)

type jwtClaims struct {
	jwt.StandardClaims
	Id int `json:"id"`
}

type hAPI struct{ *gin.RouterGroup }

func (_ *hAPI) error(c *gin.Context, code int, err error) {
	statusCode, errorCode := ucode(code)
	out := gin.H{"code": errorCode}
	if yams.Debug && err != nil {
		out["debug"] = err.Error()
	}
	c.AbortWithStatusJSON(statusCode, out)
}

func (h *hAPI) unauthorized(c *gin.Context, errorCode int, err error) {
	c.Header("www-authenticate", `Bearer realm="YAMS API"`)
	h.error(c, code(http.StatusUnauthorized, errorCode), err)
}

func (h *hAPI) forbidden(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusForbidden, errorCode), err)
}

func (h *hAPI) notFound(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusNotFound, errorCode), err)
}

func (h *hAPI) conflict(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusConflict, errorCode), err)
}

func (h *hAPI) requestEntityTooLarge(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusRequestEntityTooLarge, errorCode), err)
}

func (h *hAPI) unsupportedMediaType(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusUnsupportedMediaType, errorCode), err)
}

func (h *hAPI) ok(c *gin.Context, obj interface{}) {
	if obj != nil {
		c.AbortWithStatusJSON(http.StatusOK, obj)
	} else {
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func (_ *hAPI) paramInt(c *gin.Context, key string) int {
	v, _ := strconv.Atoi(c.Param(key))
	return v
}

func (h *hAPI) bind(c *gin.Context, obj interface{}) bool {
	var b binding.Binding = binding.Form
	if c.Request.Method != "GET" {
		switch c.ContentType() {
		case gin.MIMEJSON:
			b = binding.JSON
		case gin.MIMEXML, gin.MIMEXML2:
			b = binding.XML
		case gin.MIMEPOSTForm:
			b = binding.FormPost
		case gin.MIMEMultipartPOSTForm:
			b = binding.FormMultipart
		}
	}
	err := c.ShouldBindWith(obj, b)
	if err == nil {
		return true
	}
	if _, ok := err.(validator.ValidationErrors); ok {
		h.error(c, http.StatusUnprocessableEntity, err)
	} else {
		h.error(c, http.StatusBadRequest, err)
	}
	return false
}

func (_ *hAPI) token(claims *jwtClaims) string {
	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(yams.SecretKey)
	if err != nil {
		panic(err)
	}
	return token
}

func (h *hAPI) checkUser(c *gin.Context, id int, checkSelf bool) bool {
	if id == 0 {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return false
	}
	if checkSelf && id == c.MustGet("jwt").(*jwtClaims).Id {
		h.forbidden(c, errCodeInvalidIdentifier, nil)
		return false
	}
	q := qb.Select("id").From("users").Where("id = ?", id)
	if err := q.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return false
		}
		panic(err)
	}
	return true
}

func (h *hAPI) checkProfileAccess(c *gin.Context, id int) bool {
	if id == 0 {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return false
	}
	q := qb.Select("id").From("profiles").Where("id = ?", id)
	if err := q.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return false
		}
		panic(err)
	}
	if aclId := c.MustGet("acl_id"); aclId != nil && !yams.InIntSlice(aclId.([]int), id) {
		h.forbidden(c, errCodeAuthFailedACL, nil)
		return false
	}
	return true
}

func (h *hAPI) checkRouteAccess(c *gin.Context, id int) bool {
	if id == 0 {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return false
	}
	var pid int
	q := qb.Select("profile_id").From("routes").Where("id = ?", id)
	if err := q.Scan(&pid); err != nil {
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return false
		}
		panic(err)
	}
	if aclId := c.MustGet("acl_id"); aclId != nil && !yams.InIntSlice(aclId.([]int), pid) {
		h.forbidden(c, errCodeAuthFailedACL, nil)
		return false
	}
	return true
}

func (h *hAPI) Authentication(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		h.unauthorized(c, errCodeAuthNoHeader, nil)
		return
	}
	claims := &jwtClaims{}
	_, err := jwt.ParseWithClaims(auth[7:], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(`unexpected signing method "%v"`, token.Header["alg"])
		}
		return yams.SecretKey, nil
	})
	if err != nil {
		h.unauthorized(c, errCodeAuthBadToken, err)
		return
	}
	c.Set("jwt", claims)

	var acl []string
	q := qb.Update("users").Set("last_auth_at", sqrl.Expr("now()")).Where("id = ?", claims.Id).Suffix("RETURNING acl")
	if err = q.Scan(pq.Array(&acl)); err != nil {
		if err == sql.ErrNoRows {
			h.unauthorized(c, errCodeAuthNoUser, nil)
			return
		}
		panic(err)
	}
	c.Set("acl", acl)
	c.Set("acl_id", nil)
	c.Next()
}

func (h *hAPI) ACL(ns ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		aclMap := make(map[string][]int)
		for _, acl := range c.MustGet("acl").([]string) {
			r := strings.SplitN(acl, ":", 2)
			if v, ok := aclMap[r[0]]; len(r) == 2 {
				vc, _ := strconv.Atoi(r[1])
				aclMap[r[0]] = append(v, vc)
			} else if !ok {
				aclMap[r[0]] = v
			}
		}
		if _, ok := aclMap[aclAdmin]; ok {
			c.Next()
			return
		}
		for _, ns := range ns {
			if id, ok := aclMap[ns]; ok {
				if len(id) > 0 {
					c.Set("acl_id", id)
				}
				c.Next()
				return
			}
		}
		h.forbidden(c, errCodeAuthFailedACL, nil)
	}
}

func (h *hAPI) IndexAction(c *gin.Context) {
	c.JSON(200, gin.H{
		"uptime": time.Since(upTime).String(),
	})
}

func (h *hAPI) AuthAction(c *gin.Context) {
	var in struct {
		Username string `form:"username" json:"username" binding:"required,username"`
		Password string `form:"password" json:"password" binding:"required,min=3,max=72"`
	}
	if !h.bind(c, &in) {
		return
	}

	claims := &jwtClaims{}
	q := qb.Select("id").From("users").Where("username = ? AND password = crypt(?, password)", in.Username, in.Password)
	if err := q.Scan(&claims.Id); err != nil {
		if err == sql.ErrNoRows {
			h.unauthorized(c, errCodeBadCredentials, nil)
			return
		}
		panic(err)
	}
	h.ok(c, gin.H{"token": h.token(claims)})
}

func (h *hAPI) AuthACLAction(c *gin.Context) {
	h.ok(c, c.MustGet("acl").([]string))
}

func (h *hAPI) AuthRefreshAction(c *gin.Context) {
	h.ok(c, gin.H{"token": h.token(c.MustGet("jwt").(*jwtClaims))})
}

type outUser struct {
	Id         int            `json:"id"`
	Username   string         `json:"username"`
	ACL        pq.StringArray `json:"acl"`
	LastAuthAt *time.Time     `json:"last_auth_at"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (h *hAPI) UsersAction(c *gin.Context) {
	q := qb.Select("id", "username", "acl", "last_auth_at", "created_at").From("users").OrderBy("username")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outUser, 0)
	for rows.Next() {
		var out outUser
		if err = rows.Scan(&out.Id, &out.Username, &out.ACL, &out.LastAuthAt, &out.CreatedAt); err != nil {
			panic(err)
		}
		rs = append(rs, out)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	h.ok(c, rs)
}

func (h *hAPI) UsersViewAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkUser(c, id, false) {
		return
	}

	var out outUser
	q := qb.Select("id", "username", "acl", "last_auth_at", "created_at").From("users").Where("id = ?", id)
	if err := q.Scan(&out.Id, &out.Username, &out.ACL, &out.LastAuthAt, &out.CreatedAt); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

type inUser struct {
	Username string   `form:"username" json:"username" binding:"required,username"`
	Password string   `form:"password" json:"password" binding:"omitempty,min=3,max=72"`
	ACL      []string `form:"acl" json:"acl" binding:"unique,dive,required,max=32,acl"`
}

func (h *hAPI) UsersCreateAction(c *gin.Context) {
	var in inUser
	if !h.bind(c, &in) {
		return
	}

	q1 := qb.Select("TRUE").From("users").Where("username = lower(?)", in.Username)
	if err := q1.Scan(new(bool)); err == nil {
		h.conflict(c, errCodeUsernameExists, nil)
		return
	} else if err != sql.ErrNoRows {
		panic(err)
	}

	var out struct {
		Id       int    `json:"id"`
		Password string `json:"password,omitempty"`
	}
	if in.Password == "" {
		in.Password = yams.RandString(8)
		out.Password = in.Password
	}

	q2 := qb.Insert("users").SetMap(map[string]interface{}{
		"username": in.Username,
		"password": sqrl.Expr("crypt(?, gen_salt('bf'))", in.Password),
		"acl":      pq.Array(in.ACL),
	}).Suffix("RETURNING id")
	if err := q2.Scan(&out.Id); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

func (h *hAPI) UsersUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkUser(c, id, false) {
		return
	}

	var in inUser
	if !h.bind(c, &in) {
		return
	}

	q1 := qb.Select("TRUE").From("users").Where("id <> ? AND username = lower(?)", id, in.Username)
	if err := q1.Scan(new(bool)); err == nil {
		h.conflict(c, errCodeUsernameExists, nil)
		return
	} else if err != sql.ErrNoRows {
		panic(err)
	}

	q2 := qb.Update("users").SetMap(map[string]interface{}{
		"username": in.Username,
		"acl":      pq.Array(in.ACL),
	}).Where("id = ?", id)
	if in.Password != "" {
		q2.Set("password", sqrl.Expr("crypt(?, gen_salt('bf'))", in.Password))
	}
	if _, err := q2.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) UsersDeleteAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkUser(c, id, true) {
		return
	}

	q := qb.Delete("users").Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

type outProfile struct {
	Id           int            `json:"id"`
	Name         string         `json:"name"`
	Hosts        pq.StringArray `json:"hosts"`
	Backend      *string        `json:"backend"`
	Debug        bool           `json:"debug"`
	VarsLifetime int            `json:"vars_lifetime"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (h *hAPI) ProfilesAction(c *gin.Context) {
	q := qb.Select("id", "name", "hosts", "backend", "debug", "vars_lifetime", "created_at").From("profiles").OrderBy("name")
	if aclId := c.MustGet("acl_id"); aclId != nil {
		q.Where(sqrl.Eq{"id": aclId.([]int)})
	}
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outProfile, 0)
	for rows.Next() {
		var out outProfile
		if err = rows.Scan(&out.Id, &out.Name, &out.Hosts, &out.Backend, &out.Debug, &out.VarsLifetime, &out.CreatedAt); err != nil {
			panic(err)
		}
		rs = append(rs, out)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	h.ok(c, rs)
}

func (h *hAPI) ProfilesViewAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, id) {
		return
	}

	var out outProfile
	q := qb.Select("id", "name", "hosts", "backend", "debug", "vars_lifetime", "created_at").From("profiles").Where("id = ?", id)
	if err := q.Scan(&out.Id, &out.Name, &out.Hosts, &out.Backend, &out.Debug, &out.VarsLifetime, &out.CreatedAt); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

type inProfile struct {
	Name         string   `form:"name" json:"name" binding:"required,min=3,max=72"`
	Hosts        []string `form:"hosts" json:"hosts" binding:"required,min=1,dive,required,max=128,host"`
	Backend      *string  `form:"backend" json:"backend" binding:"omitempty,required,max=128,url"`
	Debug        bool     `form:"debug" json:"debug" binding:"omitempty"`
	VarsLifetime int      `form:"vars_lifetime" json:"vars_lifetime" binding:"omitempty,min=1,max=2147483647"`
}

func (h *hAPI) ProfilesCreateAction(c *gin.Context) {
	var in inProfile
	if !h.bind(c, &in) {
		return
	}

	q1 := qb.Select("TRUE").From("profiles").Where("hosts && ?", pq.Array(in.Hosts))
	if err := q1.Scan(new(bool)); err == nil {
		h.conflict(c, errCodeProfileHostExists, nil)
		return
	} else if err != sql.ErrNoRows {
		panic(err)
	}

	var id int
	q2 := qb.Insert("profiles").SetMap(map[string]interface{}{
		"name":          in.Name,
		"hosts":         pq.Array(in.Hosts),
		"backend":       in.Backend,
		"debug":         in.Debug,
		"vars_lifetime": in.VarsLifetime,
	}).Suffix("RETURNING id")
	if err := q2.Scan(&id); err != nil {
		panic(err)
	}
	h.ok(c, gin.H{"id": id})
}

func (h *hAPI) ProfilesUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, id) {
		return
	}

	var in inProfile
	if !h.bind(c, &in) {
		return
	}

	q1 := qb.Select("TRUE").From("profiles").Where("id <> ? AND hosts && ?", id, pq.Array(in.Hosts))
	if err := q1.Scan(new(bool)); err == nil {
		h.conflict(c, errCodeProfileHostExists, nil)
		return
	} else if err != sql.ErrNoRows {
		panic(err)
	}

	q2 := qb.Update("profiles").SetMap(map[string]interface{}{
		"name":          in.Name,
		"hosts":         pq.Array(in.Hosts),
		"backend":       in.Backend,
		"debug":         in.Debug,
		"vars_lifetime": in.VarsLifetime,
	}).Where("id = ?", id)
	if _, err := q2.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) ProfilesDeleteAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, id) {
		return
	}

	q := qb.Delete("profiles").Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

type outRoute struct {
	Id      int            `json:"id"`
	Uuid    string         `json:"uuid"`
	Methods pq.StringArray `json:"methods"`
	Path    string         `json:"path"`
	Timeout int            `json:"timeout"`
	Hint    *string        `json:"hint,omitempty"`
}

func (h *hAPI) RoutesAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	q := qb.Select("id", "uuid", "methods", "path", "timeout", "hint").From("routes").Where("profile_id = ?", pid).OrderBy("position")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outRoute, 0)
	for rows.Next() {
		var out outRoute
		if err = rows.Scan(&out.Id, &out.Uuid, &out.Methods, &out.Path, &out.Timeout, &out.Hint); err != nil {
			panic(err)
		}
		rs = append(rs, out)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	h.ok(c, rs)
}

func (h *hAPI) RoutesViewAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	var out outRoute
	q := qb.Select("id", "uuid", "methods", "path", "timeout", "hint").From("routes").Where("id = ?", id)
	if err := q.Scan(&out.Id, &out.Uuid, &out.Methods, &out.Path, &out.Timeout, &out.Hint); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

type inRoute struct {
	Methods []string `form:"methods" json:"methods" binding:"required,min=1,dive,required,max=16"`
	Path    string   `form:"path" json:"path" binding:"required,max=255,prefix=/"`
	Timeout int      `form:"timeout" json:"timeout" binding:"omitempty,min=0,max=86400"`
	Hint    *string  `form:"hint" json:"hint" binding:"omitempty,required,max=255"`
}

func (h *hAPI) RoutesCreateAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	var in inRoute
	if !h.bind(c, &in) {
		return
	}

	var id int
	q := qb.Insert("routes").SetMap(map[string]interface{}{
		"profile_id": pid,
		"methods":    pq.Array(in.Methods),
		"path":       in.Path,
		"script":     yams.DefaultScript,
		"timeout":    in.Timeout,
		"hint":       in.Hint,
	}).Suffix("RETURNING id")
	if err := q.Scan(&id); err != nil {
		panic(err)
	}
	h.ok(c, gin.H{"id": id})
}

func (h *hAPI) RoutesUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	var in inRoute
	if !h.bind(c, &in) {
		return
	}

	q := qb.Update("routes").SetMap(map[string]interface{}{
		"methods": pq.Array(in.Methods),
		"path":    in.Path,
		"timeout": in.Timeout,
		"hint":    in.Hint,
	}).Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) RoutesScriptViewAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	var adapter string
	var script sql.RawBytes
	q := qb.Select("adapter", "script").From("routes").Where("id = ?", id)
	rows, err := q.Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	rows.Next()
	if err = rows.Scan(&adapter, &script); err != nil {
		panic(err)
	}
	c.Data(200, yams.Adapters.GetMimeType(adapter), script)
}

func (h *hAPI) RoutesScriptUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	adapter, ok := yams.Adapters[c.ContentType()]
	if !ok {
		h.unsupportedMediaType(c, errCodeInvalidAdapter, nil)
		return
	}

	if c.Request.ContentLength > yams.MaxScriptSize {
		h.requestEntityTooLarge(c, errCodeUnknown, nil)
		return
	}

	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	q := qb.Update("routes").SetMap(map[string]interface{}{
		"adapter": adapter,
		"script":  buf,
	}).Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) RoutesPositionAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	var in struct {
		Position int `form:"position" json:"position" binding:"omitempty,min=0"`
	}
	if !h.bind(c, &in) {
		return
	}

	q := qb.Update("routes").Set("position", in.Position).Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) RoutesDeleteAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	q := qb.Delete("routes").Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

type outAsset struct {
	Id        int       `json:"id"`
	Path      string    `json:"path"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *hAPI) AssetsAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	q := qb.Select("id", "path", "mime_type", "created_at").From("assets").Where("profile_id = ?", pid).OrderBy("path")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outAsset, 0)
	for rows.Next() {
		var out outAsset
		if err = rows.Scan(&out.Id, &out.Path, &out.MimeType, &out.CreatedAt); err != nil {
			panic(err)
		}
		rs = append(rs, out)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	h.ok(c, rs)
}

func (h *hAPI) AssetsUploadAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	mimeType := c.GetHeader("Content-Type")
	if mimeType == "" {
		h.unsupportedMediaType(c, errCodeUnknown, nil)
		return
	}

	if c.Request.ContentLength > yams.MaxAssetSize {
		h.requestEntityTooLarge(c, errCodeUnknown, nil)
		return
	}

	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	values := map[string]interface{}{
		"profile_id": pid,
		"data":       buf,
		"mime_type":  mimeType,
	}
	p := strings.Trim(c.Param("path"), "/")
	if p != "" {
		values["path"] = p
	}

	q := qb.Insert("assets").SetMap(values)
	q.Suffix(`ON CONFLICT (profile_id, path) DO UPDATE SET data = EXCLUDED.data, mime_type = EXCLUDED.mime_type, created_at = DEFAULT RETURNING path`)
	if err := q.Scan(&p); err != nil {
		panic(err)
	}
	h.ok(c, gin.H{"path": p})
}

func (h *hAPI) AssetsDownloadAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	var data sql.RawBytes
	var mimeType string
	p := c.Param("path")[1:]

	// Writes asset data directly to Writer.
	q := qb.Select("data", "mime_type").From("assets").Where("profile_id = ? AND path = ?", pid, p)
	rows, err := q.Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if !rows.Next() {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
	}
	if err = rows.Scan(&data, &mimeType); err != nil {
		panic(err)
	}
	c.Data(200, mimeType, data)
}

func (h *hAPI) AssetsDeleteAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	p := c.Param("path")[1:]
	q := qb.Delete("assets").Where("profile_id = ? AND path = ?", pid, p).Suffix("RETURNING TRUE")
	if err := q.Scan(new(bool)); err != nil {
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return
		}
		panic(err)
	}
	h.ok(c, nil)
}
