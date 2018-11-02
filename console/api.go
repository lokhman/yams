package console

import (
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	errCodeInvalidAdapter
)

type hAPI struct{ *gin.RouterGroup }

func (_ *hAPI) error(c *gin.Context, code int, err error, vars gin.H) {
	statusCode, errorCode := ucode(code)
	out := gin.H{"code": errorCode}
	if yams.Debug && err != nil {
		out["debug"] = err.Error()
	}
	for k, v := range vars {
		out[k] = v
	}
	io.Copy(ioutil.Discard, c.Request.Body)
	c.AbortWithStatusJSON(statusCode, out)
}

func (h *hAPI) unauthorized(c *gin.Context, errorCode int, err error) {
	c.Header("WWW-Authenticate", `Bearer realm="YAMS API"`)
	h.error(c, code(http.StatusUnauthorized, errorCode), err, nil)
}

func (h *hAPI) forbidden(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusForbidden, errorCode), err, nil)
}

func (h *hAPI) notFound(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusNotFound, errorCode), err, nil)
}

func (h *hAPI) conflict(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusConflict, errorCode), err, nil)
}

func (h *hAPI) requestEntityTooLarge(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusRequestEntityTooLarge, errorCode), err, nil)
}

func (h *hAPI) unsupportedMediaType(c *gin.Context, errorCode int, err error) {
	h.error(c, code(http.StatusUnsupportedMediaType, errorCode), err, nil)
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
	return h.checkValid(c, c.ShouldBindWith(obj, b))
}

func (h *hAPI) checkValid(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	if ves, ok := err.(validator.ValidationErrors); ok {
		vars := gin.H{}
		for _, ve := range ves {
			vars[ve.Field()] = ve.Tag()
		}
		h.error(c, http.StatusUnprocessableEntity, err, gin.H{"invalid": vars})
	} else {
		h.error(c, code(http.StatusBadRequest, errCodeUnknown), err, nil)
	}
	return false
}

func (h *hAPI) checkACLAccess(c *gin.Context, id int) bool {
	if acl := c.MustGet("acl").([]int64); acl != nil {
		for _, v := range acl {
			if v == int64(id) {
				return true
			}
		}
		h.forbidden(c, errCodeAuthFailedACL, nil)
		return false
	}
	return true
}

func (h *hAPI) checkUserAccess(c *gin.Context, id int, checkSelf bool) bool {
	if id == 0 {
		h.notFound(c, errCodeInvalidIdentifier, nil)
		return false
	}
	if checkSelf && id == c.MustGet("auth").(yams.JWTClaims).Id {
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
	return h.checkACLAccess(c, id)
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
	return h.checkACLAccess(c, pid)
}

func (h *hAPI) getToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		return token[7:]
	}
	token, _ = c.Cookie("token")
	return token
}

type outUser struct {
	Id         int           `json:"id"`
	Username   string        `json:"username"`
	Role       string        `json:"role"`
	ACL        pq.Int64Array `json:"acl,omitempty"`
	LastAuthAt *time.Time    `json:"last_auth_at"`
	CreatedAt  time.Time     `json:"created_at"`
}

func (h *hAPI) outUser(c *gin.Context, id int) {
	var out outUser
	q := qb.Select("id", "username", "role", "array_remove(array_agg(profile_id), NULL)", "last_auth_at", "created_at").
		From("users").LeftJoin("acl ON user_id = id").Where("id = ?", id).GroupBy("id")
	if err := q.Scan(&out.Id, &out.Username, &out.Role, &out.ACL, &out.LastAuthAt, &out.CreatedAt); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

func (h *hAPI) Auth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := h.getToken(c)
		if token == "" {
			h.unauthorized(c, errCodeAuthNoHeader, nil)
			return
		}
		auth := yams.JWTClaims{}
		if _, err := yams.JWTParse(token, &auth); err != nil {
			h.unauthorized(c, errCodeAuthBadToken, err)
			return
		}

		var acl []int64
		q := qb.Select("username", "role", "array_remove(array_agg(profile_id), NULL)").From("users").
			LeftJoin("acl ON user_id = id").Where("id = ? AND role = ANY(?)", auth.Id, pq.StringArray(roles)).GroupBy("id")
		if err := q.Scan(&auth.Username, &auth.Role, pq.Array(&acl)); err != nil {
			if err == sql.ErrNoRows {
				h.forbidden(c, errCodeAuthNoUser, nil)
				return
			}
			panic(err)
		}

		switch auth.Role {
		case yams.RoleManager, yams.RoleAdmin:
			if len(acl) == 0 {
				acl = nil
			}
		}

		c.Set("auth", auth)
		c.Set("acl", acl)
		c.Next()
	}
}

func (h *hAPI) IndexAction(c *gin.Context) {
	c.JSON(200, gin.H{
		"uptime": time.Since(upTime).String(),
	})
}

func (h *hAPI) AuthAction(c *gin.Context) {
	var in struct {
		Username string `form:"username" json:"username" binding:"required,trim,username"`
		Password string `form:"password" json:"password" binding:"required,trim,min=3,max=72"`
	}
	if !h.bind(c, &in) {
		return
	}

	auth := yams.JWTClaims{}
	q := qb.Update("users").Set("last_auth_at", sqrl.Expr("now()")).
		Where("username = ? AND password = crypt(?, password)", in.Username, in.Password).
		Suffix("RETURNING id")
	if err := q.Scan(&auth.Id); err != nil {
		if err == sql.ErrNoRows {
			h.unauthorized(c, errCodeBadCredentials, nil)
			return
		}
		panic(err)
	}

	h.ok(c, gin.H{"token": yams.JWTSign(auth)})
}

func (h *hAPI) AuthRefreshAction(c *gin.Context) {
	h.ok(c, gin.H{"token": yams.JWTSign(c.MustGet("auth").(yams.JWTClaims))})
}

func (h *hAPI) AuthPasswordAction(c *gin.Context) {
	var in struct {
		Old string `form:"old" json:"old" binding:"required,trim,min=3,max=72"`
		New string `form:"new" json:"new" binding:"required,trim,min=3,max=72,nefield=Old"`
	}
	if !h.bind(c, &in) {
		return
	}

	id := c.MustGet("auth").(yams.JWTClaims).Id
	q := qb.Update("users").Set("password", sqrl.Expr("crypt(?, gen_salt('bf'))", in.New)).
		Where("id = ? AND password = crypt(?, password)", id, in.Old).
		Suffix("RETURNING TRUE")
	if err := q.Scan(new(bool)); err != nil {
		if err == sql.ErrNoRows {
			h.forbidden(c, errCodeBadCredentials, nil)
			return
		}
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) AuthUserAction(c *gin.Context) {
	h.outUser(c, c.MustGet("auth").(yams.JWTClaims).Id)
}

func (h *hAPI) UsersAction(c *gin.Context) {
	q := qb.Select("id", "username", "role", "array_remove(array_agg(profile_id), NULL)", "last_auth_at", "created_at").
		From("users").LeftJoin("acl ON user_id = id").GroupBy("id").OrderBy("username")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outUser, 0)
	for rows.Next() {
		var out outUser
		if err = rows.Scan(&out.Id, &out.Username, &out.Role, &out.ACL, &out.LastAuthAt, &out.CreatedAt); err != nil {
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
	if !h.checkUserAccess(c, id, false) {
		return
	}

	h.outUser(c, id)
}

type inUser struct {
	id       int
	Username string `form:"username" json:"username" binding:"required,trim,username"`
	Password string `form:"password" json:"password" binding:"omitempty,trim,min=3,max=72"`
	Role     string `form:"role" json:"role" binding:"required,trim,role"`
	ACL      []int  `form:"acl" json:"acl" binding:"unique,acl"`
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

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	q2 := qb.Insert("users").SetMap(gin.H{
		"username": in.Username,
		"password": sqrl.Expr("crypt(?, gen_salt('bf'))", in.Password),
		"role":     in.Role,
	}).Suffix("RETURNING id")
	if err = q2.RunWith(tx).Scan(&out.Id); err != nil {
		panic(err)
	}

	if len(in.ACL) > 0 {
		q := qb.Insert("acl").Columns("user_id", "profile_id")
		for _, pid := range in.ACL {
			q.Values(out.Id, pid)
		}
		if _, err = q.RunWith(tx).Exec(); err != nil {
			panic(err)
		}
	}

	if err = tx.Commit(); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

func (h *hAPI) UsersUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkUserAccess(c, id, false) {
		return
	}

	in := inUser{id: id}
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

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	q2 := qb.Update("users").SetMap(gin.H{
		"username": in.Username,
		"role":     in.Role,
	}).Where("id = ?", id)
	if in.Password != "" {
		q2.Set("password", sqrl.Expr("crypt(?, gen_salt('bf'))", in.Password))
	}
	if _, err := q2.RunWith(tx).Exec(); err != nil {
		panic(err)
	}

	q3 := qb.Delete("acl").Where("user_id = ?", id)
	if _, err := q3.RunWith(tx).Exec(); err != nil {
		panic(err)
	}
	if len(in.ACL) > 0 {
		q := qb.Insert("acl").Columns("user_id", "profile_id")
		for _, pid := range in.ACL {
			q.Values(id, pid)
		}
		if _, err = q.RunWith(tx).Exec(); err != nil {
			panic(err)
		}
	}

	if err = tx.Commit(); err != nil {
		panic(err)
	}
	h.ok(c, nil)
}

func (h *hAPI) UsersDeleteAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkUserAccess(c, id, true) {
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
	IsDebug      bool           `json:"is_debug"`
	VarsLifetime int            `json:"vars_lifetime"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (h *hAPI) ProfilesAction(c *gin.Context) {
	if preview, _ := strconv.ParseBool(c.Query("preview")); preview {
		h.ProfilesPreviewAction(c)
		return
	}

	q := qb.Select("id", "name", "hosts", "backend", "is_debug", "vars_lifetime", "created_at").From("profiles").OrderBy("name")
	if acl := c.MustGet("acl").([]int64); acl != nil {
		q.Where(sqrl.Eq{"id": acl})
	}
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outProfile, 0)
	for rows.Next() {
		var out outProfile
		if err = rows.Scan(&out.Id, &out.Name, &out.Hosts, &out.Backend, &out.IsDebug, &out.VarsLifetime, &out.CreatedAt); err != nil {
			panic(err)
		}
		rs = append(rs, out)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	h.ok(c, rs)
}

func (h *hAPI) ProfilesPreviewAction(c *gin.Context) {
	type out struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	q := qb.Select("id", "name").From("profiles").OrderBy("name")
	if acl := c.MustGet("acl").([]int64); acl != nil {
		q.Where(sqrl.Eq{"id": acl})
	}
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]out, 0)
	for rows.Next() {
		var out out
		if err = rows.Scan(&out.Id, &out.Name); err != nil {
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
	q := qb.Select("id", "name", "hosts", "backend", "is_debug", "vars_lifetime", "created_at").From("profiles").Where("id = ?", id)
	if err := q.Scan(&out.Id, &out.Name, &out.Hosts, &out.Backend, &out.IsDebug, &out.VarsLifetime, &out.CreatedAt); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

type inProfile struct {
	id           int
	Name         string   `form:"name" json:"name" binding:"required,trim,min=3,max=72"`
	Hosts        []string `form:"hosts" json:"hosts" binding:"required,min=1,dive,required,trim,max=128,host"`
	Backend      *string  `form:"backend" json:"backend" binding:"omitempty,required,trim,max=128,url"`
	IsDebug      bool     `form:"is_debug" json:"is_debug" binding:"omitempty"`
	VarsLifetime int      `form:"vars_lifetime" json:"vars_lifetime" binding:"required,min=1,max=2147483647"`
}

func (h *hAPI) ProfilesCreateAction(c *gin.Context) {
	var in inProfile
	if !h.bind(c, &in) {
		return
	}

	var id int
	q := qb.Insert("profiles").SetMap(gin.H{
		"name":          in.Name,
		"hosts":         pq.StringArray(in.Hosts),
		"backend":       in.Backend,
		"is_debug":      in.IsDebug,
		"vars_lifetime": in.VarsLifetime,
	}).Suffix("RETURNING id")
	if err := q.Scan(&id); err != nil {
		panic(err)
	}
	h.ok(c, gin.H{"id": id})
}

func (h *hAPI) ProfilesUpdateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, id) {
		return
	}

	in := inProfile{id: id}
	if !h.bind(c, &in) {
		return
	}

	q := qb.Update("profiles").SetMap(gin.H{
		"name":          in.Name,
		"hosts":         pq.StringArray(in.Hosts),
		"backend":       in.Backend,
		"is_debug":      in.IsDebug,
		"vars_lifetime": in.VarsLifetime,
	}).Where("id = ?", id)
	if _, err := q.Exec(); err != nil {
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
	Id         int            `json:"id"`
	Uuid       string         `json:"uuid"`
	Path       string         `json:"path"`
	Methods    pq.StringArray `json:"methods"`
	Adapter    string         `json:"adapter"`
	ScriptSize int            `json:"script_size"`
	Timeout    int            `json:"timeout"`
	Hint       *string        `json:"hint,omitempty"`
	IsEnabled  bool           `json:"is_enabled"`
}

func (h *hAPI) RoutesAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	q := qb.Select("id", "uuid", "path", "methods", "adapter", "octet_length(script)", "timeout", "hint", "is_enabled").
		From("routes").Where("profile_id = ?", pid).OrderBy("position")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outRoute, 0)
	for rows.Next() {
		var out outRoute
		if err = rows.Scan(&out.Id, &out.Uuid, &out.Path, &out.Methods, &out.Adapter, &out.ScriptSize, &out.Timeout, &out.Hint, &out.IsEnabled); err != nil {
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
	q := qb.Select("id", "uuid", "path", "methods", "adapter", "octet_length(script)", "timeout", "hint", "is_enabled").From("routes").Where("id = ?", id)
	if err := q.Scan(&out.Id, &out.Uuid, &out.Path, &out.Methods, &out.Adapter, &out.ScriptSize, &out.Timeout, &out.Hint, &out.IsEnabled); err != nil {
		panic(err)
	}
	h.ok(c, out)
}

type inRoute struct {
	id        int
	Path      string   `form:"path" json:"path" binding:"required,trim,max=255,prefix=/"`
	Methods   []string `form:"methods" json:"methods" binding:"required,min=1,dive,required,trim,max=16"`
	Timeout   int      `form:"timeout" json:"timeout" binding:"omitempty,min=0,max=86400"`
	Hint      *string  `form:"hint" json:"hint" binding:"omitempty,required,trim,max=255"`
	IsEnabled bool     `form:"is_enabled" json:"is_enabled" binding:"omitempty"`
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
	q := qb.Insert("routes").SetMap(gin.H{
		"profile_id": pid,
		"path":       in.Path,
		"methods":    pq.StringArray(in.Methods),
		"script":     yams.DefaultScript,
		"timeout":    in.Timeout,
		"hint":       in.Hint,
		"is_enabled": in.IsEnabled,
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

	in := inRoute{id: id}
	if !h.bind(c, &in) {
		return
	}

	q := qb.Update("routes").SetMap(gin.H{
		"path":       in.Path,
		"methods":    pq.StringArray(in.Methods),
		"timeout":    in.Timeout,
		"hint":       in.Hint,
		"is_enabled": in.IsEnabled,
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

	q := qb.Update("routes").SetMap(gin.H{
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

func (h *hAPI) RoutesStateAction(c *gin.Context) {
	id := h.paramInt(c, "id")
	if !h.checkRouteAccess(c, id) {
		return
	}

	var in struct {
		IsEnabled bool `form:"is_enabled" json:"is_enabled" binding:"omitempty"`
	}
	if !h.bind(c, &in) {
		return
	}

	q := qb.Update("routes").Set("is_enabled", in.IsEnabled).Where("id = ?", id)
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
	Path      string    `json:"path"`
	MimeType  string    `json:"mime_type"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *hAPI) AssetsAction(c *gin.Context) {
	pid := h.paramInt(c, "id")
	if !h.checkProfileAccess(c, pid) {
		return
	}

	q := qb.Select("path", "mime_type", "octet_length(data)", "created_at").From("assets").Where("profile_id = ?", pid).OrderBy("path")
	rows, err := q.Query()
	defer rows.Close()

	rs := make([]outAsset, 0)
	for rows.Next() {
		var out outAsset
		if err = rows.Scan(&out.Path, &out.MimeType, &out.Size, &out.CreatedAt); err != nil {
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

	values := gin.H{
		"profile_id": pid,
		"data":       buf,
		"mime_type":  mimeType,
	}
	p := strings.Trim(c.Param("path"), "/")
	if p != "" && len(p) <= 72 {
		values["path"] = p
	}

	q := qb.Insert("assets").SetMap(values)
	q.Suffix("ON CONFLICT (profile_id, path) DO UPDATE SET data = EXCLUDED.data, mime_type = EXCLUDED.mime_type, created_at = DEFAULT RETURNING path")
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
