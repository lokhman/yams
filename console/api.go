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
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
	"github.com/lokhman/sqrl"
	"github.com/lokhman/yams/yams"
	"gopkg.in/go-playground/validator.v9"
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

func (h *hAPI) checkProfileAccess(c *gin.Context, id int) bool {
	var aclId interface{}
	if id == 0 {
		goto deny
	}
	aclId = c.MustGet("acl_id")
	if aclId != nil && !yams.InIntSlice(aclId.([]int), id) {
		goto deny
	}
	return true
deny:
	h.notFound(c, errCodeInvalidIdentifier, nil)
	return false
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

	var aclList pq.StringArray
	q := qb.Select("acl").From("users").Where("id = ?", claims.Id)
	if err = q.Scan(&aclList); err != nil {
		if err == sql.ErrNoRows {
			h.unauthorized(c, errCodeAuthNoUser, nil)
			return
		}
		panic(err)
	}
	aclMap := make(map[string][]int)
	for _, acl := range aclList {
		r := strings.SplitN(acl, ":", 2)
		if v, ok := aclMap[r[0]]; len(r) == 2 {
			vc, _ := strconv.Atoi(r[1])
			aclMap[r[0]] = append(v, vc)
		} else if !ok {
			aclMap[r[0]] = v
		}
	}
	c.Set("acl", aclMap)
	c.Set("acl_id", nil)
	c.Next()
}

func (h *hAPI) ACL(ns ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		acl := c.MustGet("acl").(map[string][]int)
		if _, ok := acl[aclAdmin]; ok {
			c.Next()
			return
		}
		for _, ns := range ns {
			if id, ok := acl[ns]; ok {
				if len(id) > 0 {
					c.Set("acl_id", id)
				}
				c.Next()
				return
			}
		}
		h.forbidden(c, errCodeAuthNoACL, nil)
	}
}

func (h *hAPI) IndexAction(c *gin.Context) {
	c.JSON(200, gin.H{
		"uptime": time.Since(upTime).String(),
	})
}

func (h *hAPI) AuthAction(c *gin.Context) {
	var in struct {
		Username string `form:"username" json:"username" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
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
	h.ok(c, c.MustGet("acl").(map[string][]int))
}

func (h *hAPI) AuthRefreshAction(c *gin.Context) {
	h.ok(c, gin.H{"token": h.token(c.MustGet("jwt").(*jwtClaims))})
}

type outProfile struct {
	Id           int            `json:"id"`
	Name         string         `json:"name"`
	Hosts        pq.StringArray `json:"hosts"`
	Backend      *string        `json:"backend,omitempty"`
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
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return
		}
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

	var exists bool
	q1 := qb.Select("COUNT(*) > 0").From("profiles").Where("hosts && ?", pq.Array(in.Hosts))
	if err := q1.Scan(&exists); err != nil {
		panic(err)
	}
	if exists {
		h.conflict(c, errCodeProfileHostExists, nil)
		return
	}

	var id int
	q2 := qb.Insert("profiles").SetMap(map[string]interface{}{
		"name":          in.Name,
		"hosts":         pq.Array(in.Hosts),
		"backend":       in.Backend,
		"debug":         in.Debug,
		"vars_lifetime": in.VarsLifetime,
	}).Suffix("RETURNING TRUE")
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

	var exists bool
	q1 := qb.Select("COUNT(*) > 0").From("profiles").Where("id <> ? AND hosts && ?", id, pq.Array(in.Hosts))
	if err := q1.Scan(&exists); err != nil {
		panic(err)
	}
	if exists {
		h.conflict(c, errCodeProfileHostExists, nil)
		return
	}

	q2 := qb.Update("profiles").SetMap(map[string]interface{}{
		"name":          in.Name,
		"hosts":         pq.Array(in.Hosts),
		"backend":       in.Backend,
		"debug":         in.Debug,
		"vars_lifetime": in.VarsLifetime,
	}).Where("id = ?", id).Suffix("RETURNING TRUE")
	if err := q2.Scan(&exists); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if !exists {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
	}
	h.ok(c, nil)
}

func (h *hAPI) ProfilesDeleteAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, id) {
		return
	}

	var exists bool
	q := qb.Delete("profiles").Where("id = ?", id).Suffix("RETURNING TRUE")
	if err := q.Scan(&exists); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if !exists {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
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
	var out outRoute
	q := qb.Select("id", "uuid", "methods", "path", "timeout", "hint").From("routes").Where("id = ?", h.paramInt(c, "id"))
	if aclId := c.MustGet("acl_id"); aclId != nil {
		q.Where(sqrl.Eq{"profile_id": aclId.([]int)})
	}
	if err := q.Scan(&out.Id, &out.Uuid, &out.Methods, &out.Path, &out.Timeout, &out.Hint); err != nil {
		if err == sql.ErrNoRows {
			h.notFound(c, errCodeInvalidIdentifier, nil)
			return
		}
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
	var in inRoute
	if !h.bind(c, &in) {
		return
	}

	var exists bool
	q := qb.Update("routes").SetMap(map[string]interface{}{
		"methods": pq.Array(in.Methods),
		"path":    in.Path,
		"timeout": in.Timeout,
		"hint":    in.Hint,
	}).Where("id = ?", h.paramInt(c, "id")).Suffix("RETURNING TRUE")
	if aclId := c.MustGet("acl_id"); aclId != nil {
		q.Where(sqrl.Eq{"profile_id": aclId.([]int)})
	}
	if err := q.Scan(&exists); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if !exists {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
	}
	h.ok(c, nil)
}

func (h *hAPI) RoutesDeleteAction(c *gin.Context) {
	var exists bool
	q := qb.Delete("routes").Where("id = ?", h.paramInt(c, "id")).Suffix("RETURNING TRUE")
	if aclId := c.MustGet("acl_id"); aclId != nil {
		q.Where(sqrl.Eq{"profile_id": aclId.([]int)})
	}
	if err := q.Scan(&exists); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if !exists {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
	}
	h.ok(c, nil)
}

func (h *hAPI) RoutesScriptAction(c *gin.Context) {
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

	var exists bool
	q := qb.Update("routes").SetMap(map[string]interface{}{
		"adapter": adapter,
		"script":  buf,
	}).Where("id = ?", h.paramInt(c, "id")).Suffix("RETURNING TRUE")
	if aclId := c.MustGet("acl_id"); aclId != nil {
		q.Where(sqrl.Eq{"profile_id": aclId.([]int)})
	}
	if err := q.Scan(&exists); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if !exists {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return
	}
	h.ok(c, nil)
}
