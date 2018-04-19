package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/lokhman/yams-lua"
	luaBase64 "github.com/lokhman/yams-lua-base64"
	luaJSON "github.com/lokhman/yams-lua-json"
	"github.com/lokhman/yams/utils"
)

type script struct {
	route  *route
	rw     http.ResponseWriter
	req    *http.Request
	sid    string
	status int
	wbuf   []func(w http.ResponseWriter)
}

func newScript(r *route, rw http.ResponseWriter, req *http.Request, sid string) *script {
	return &script{r, rw, req, sid, 0, nil}
}

func (s *script) execute() error {
	l := lua.NewState()
	defer l.Close()

	luaJSON.Preload(l)
	luaBase64.Preload(l)
	l.PreloadModule("yams", s.loader)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.route.timeout)*time.Second)
	defer cancel()

	l.SetContext(ctx)
	if err := l.DoString(s.route.script); err != nil {
		return err
	}

	if s.status != 0 {
		s.rw.WriteHeader(s.status)
	}
	for _, buf := range s.wbuf {
		buf(s.rw)
	}
	return nil
}

func (s *script) loader(l *lua.LState) int {
	mod := l.NewTable()

	// constants
	l.SetField(mod, "routeid", lua.LString(s.route.uuid))
	l.SetField(mod, "method", lua.LString(s.req.Method))
	l.SetField(mod, "host", lua.LString(s.req.Host))
	l.SetField(mod, "uri", lua.LString(s.req.URL.Path))
	l.SetField(mod, "querystring", lua.LString(s.req.URL.RawQuery))
	l.SetField(mod, "remoteaddr", lua.LString(s.req.RemoteAddr))
	l.SetField(mod, "sessionid", lua.LString(s.sid))

	// request path parameters
	p := l.CreateTable(0, len(s.route.params))
	for k, v := range s.route.params {
		p.RawSetString(k, lua.LString(v))
	}
	l.SetField(mod, "params", p)

	// request headers
	h := l.CreateTable(0, len(s.req.Header))
	for k, vv := range s.req.Header {
		hv := l.CreateTable(len(vv), 0)
		for _, v := range vv {
			hv.Append(lua.LString(v))
		}
		h.RawSetString(k, hv)
	}
	l.SetField(mod, "headers", h)

	// request query parameters
	query := s.req.URL.Query()
	q := l.CreateTable(0, len(query))
	for k, vv := range query {
		qv := l.CreateTable(len(vv), 0)
		for _, v := range vv {
			qv.Append(lua.LString(v))
		}
		q.RawSetString(k, qv)
	}
	l.SetField(mod, "query", q)

	// request form parameters
	s.req.ParseForm()
	f := l.CreateTable(0, len(s.req.PostForm))
	for k, vv := range s.req.PostForm {
		fv := l.CreateTable(len(vv), 0)
		for _, v := range vv {
			fv.Append(lua.LString(v))
		}
		f.RawSetString(k, fv)
	}
	l.SetField(mod, "form", f)

	// exposed functions
	l.SetFuncs(mod, map[string]lua.LGFunction{
		"setstatus": s.fnSetStatus,
		"getheader": s.fnGetHeader,
		"setheader": s.fnSetHeader,
		"get":       s.fnGet,
		"asset":     s.fnAsset,
		"sleep":     s.fnSleep,
		"write":     s.fnWrite,
		"getvar":    s.fnGetVar,
		"setvar":    s.fnSetVar,
		"pass":      s.fnPass,
	})

	// register asset type
	mt := l.NewTypeMetatable(lAssetClass)
	mt.RawSetString("__index", mt)
	l.SetFuncs(mt, map[string]lua.LGFunction{
		"getmimetype": assetFnGetMimeType,
		"getsize":     assetFnGetSize,
		"template":    assetFnTemplate,
		"__tostring":  assetFnToString,
	})

	l.Push(mod)
	return 1
}

func (s *script) fnSetStatus(l *lua.LState) int {
	s.status = int(l.CheckNumber(1))
	return 0
}

func (s *script) fnGetHeader(l *lua.LState) int {
	k := l.CheckString(1)
	l.Push(lua.LString(s.req.Header.Get(k)))
	return 1
}

func (s *script) fnSetHeader(l *lua.LState) int {
	k, v := l.CheckString(1), l.CheckString(2)
	s.rw.Header().Set(k, v)
	for i := 3; i <= l.GetTop(); i++ {
		v = l.CheckString(i)
		s.rw.Header().Add(k, v)
	}
	return 0
}

func (s *script) fnGet(l *lua.LState) int {
	k := l.CheckString(1)
	if v, ok := s.req.URL.Query()[k]; ok && len(v) > 0 {
		l.Push(lua.LString(v[0]))
	} else if v, ok := s.route.params[k]; ok {
		l.Push(lua.LString(v))
	} else if v, ok := s.req.PostForm[k]; ok && len(v) > 0 {
		l.Push(lua.LString(v[0]))
	} else {
		l.Push(lua.LNil)
	}
	return 1
}

func (s *script) fnAsset(l *lua.LState) int {
	a := &asset{path: l.CheckString(1)}
	q := `SELECT id, mime_type, octet_length(data) FROM assets WHERE profile_id = $1 AND path = $2`
	if err := DB.QueryRow(q, s.route.profile.id, a.path).Scan(&a.id, &a.mimeType, &a.size); err != nil {
		if err == sql.ErrNoRows {
			l.Push(lua.LNil)
			return 1
		}
		panic(err)
	}
	ud := l.NewUserData()
	ud.Value = a
	l.SetMetatable(ud, l.GetTypeMetatable(lAssetClass))
	l.Push(ud)
	return 1
}

func (s *script) fnSleep(l *lua.LState) int {
	time.Sleep(time.Duration(l.CheckNumber(1)) * time.Second)
	return 0
}

func (s *script) fnWrite(l *lua.LState) int {
	for i := 1; i <= l.GetTop(); i++ {
		v := l.Get(i)
		if ud, ok := v.(*lua.LUserData); ok {
			switch v := ud.Value.(type) {
			case *asset:
				s.wbuf = append(s.wbuf, func(w http.ResponseWriter) {
					assetLoad(w, v)
				})
				continue
			}
		}
		s.wbuf = append(s.wbuf, func(w http.ResponseWriter) {
			fmt.Fprint(w, v)
		})
	}
	return 0
}

func (s *script) fnGetVar(l *lua.LState) int {
	k := l.CheckString(1)
	var sid *string
	if l.OptBool(2, false) {
		sid = &s.sid
	}
	var vb []byte
	q := `SELECT value FROM storage WHERE profile_id = $1 AND sid IS NOT DISTINCT FROM $2 AND key = $3 AND expires_at > now()`
	if err := DB.QueryRow(q, s.route.profile.id, sid, k).Scan(&vb); err != nil {
		if err == sql.ErrNoRows {
			l.Push(lua.LNil)
			return 1
		}
		panic(err)
	}
	q = `UPDATE storage SET expires_at = now()+(expires_at-updated_at) WHERE profile_id = $1 AND sid IS NOT DISTINCT FROM $2 AND key = $3`
	if _, err := DB.Exec(q, s.route.profile.id, sid, k); err != nil {
		panic(err)
	}
	v, err := luaJSON.Decode(l, vb)
	if err != nil {
		panic(err)
	}
	l.Push(v)
	return 1
}

func (s *script) fnSetVar(l *lua.LState) int {
	k := strings.TrimSpace(l.CheckString(1))
	if k == "" || len(k) > 255 {
		l.ArgError(1, "key must be a string of valid length [1:255]")
	}
	var sid *string
	if l.OptBool(3, false) {
		sid = &s.sid
	}
	if v := l.CheckAny(2); v != lua.LNil {
		lt := l.OptInt(4, s.route.profile.varsLft)
		if lt > s.route.profile.varsLft {
			l.ArgError(4, fmt.Sprintf("lifetime must not exceed profile setting [%d]", s.route.profile.varsLft))
		}
		vb, err := luaJSON.Encode(v)
		if err != nil {
			panic(err)
		}
		q := `INSERT INTO storage (profile_id, sid, key, value, expires_at) VALUES($1, $2, $3, $4, now() + $5 * INTERVAL '1 second') ON CONFLICT (COALESCE(sid, ''), profile_id, key) DO UPDATE SET value = EXCLUDED.value, expires_at = EXCLUDED.expires_at`
		if _, err = DB.Exec(q, s.route.profile.id, sid, k, vb, lt); err != nil {
			panic(err)
		}
	} else {
		q := `DELETE FROM storage WHERE profile_id = $1 AND sid IS NOT DISTINCT FROM $2 AND key = $3`
		if _, err := DB.Exec(q, s.route.profile.id, sid, k); err != nil {
			panic(err)
		}
	}
	return 0
}

func (s *script) fnPass(l *lua.LState) int {
	backend, target := s.route.profile.backend, ""
	if backend != nil {
		target = l.OptString(1, *backend)
	} else {
		target = l.CheckString(1)
	}
	s.route.profile.backend = &target
	defer func() { s.route.profile.backend = backend }()
	s.route.profile.proxy(s.rw, s.req)
	s.status, s.wbuf = 0, nil
	l.Exit()
	return 0
}

const lAssetClass = "ASSET*"

type asset struct {
	id       int
	path     string
	mimeType string
	size     int
}

func assetCheck(l *lua.LState) *asset {
	ud := l.CheckUserData(1)
	if v, ok := ud.Value.(*asset); ok {
		return v
	}
	l.ArgError(1, "asset expected")
	return nil
}

func assetLoad(w io.Writer, a *asset) {
	var data sql.RawBytes
	rows, err := DB.Query(`SELECT data FROM assets WHERE id = $1`, a.id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return
	}
	if err = rows.Scan(&data); err != nil {
		panic(err)
	}
	w.Write(data)
}

func assetRead(a *asset) []byte {
	buf := bytes.NewBuffer(nil)
	assetLoad(buf, a)
	return buf.Bytes()
}

func assetFnGetMimeType(l *lua.LState) int {
	l.Push(lua.LString(assetCheck(l).mimeType))
	return 1
}

func assetFnGetSize(l *lua.LState) int {
	l.Push(lua.LNumber(assetCheck(l).size))
	return 1
}

func assetFnToString(l *lua.LState) int {
	buf := assetRead(assetCheck(l))
	if s := string(buf); utils.IsBinaryString(s) {
		l.Push(lua.LString(base64.StdEncoding.EncodeToString(buf)))
	} else {
		l.Push(lua.LString(s))
	}
	return 1
}

func assetFnTemplate(l *lua.LState) int {
	asset, data := assetCheck(l), l.CheckTable(2)
	s := string(assetRead(asset))
	if utils.IsBinaryString(s) {
		l.RaiseError("template() function is not available for binary assets")
	}
	buf := bytes.NewBuffer(nil)
	t := template.Must(template.New(asset.path).Parse(s))
	if err := t.Execute(buf, scriptValueMarshal(data)); err != nil {
		panic(err)
	}
	l.Push(lua.LString(buf.Bytes()))
	return 1
}

var (
	errScriptMarshalFunction = errors.New("cannot marshal function")
	errScriptMarshalChannel  = errors.New("cannot marshal channel")
	errScriptMarshalState    = errors.New("cannot marshal state")
	errScriptMarshalUserData = errors.New("cannot marshal userdata")
	errScriptMarshalNested   = errors.New("cannot marshal recursively nested tables")
)

type scriptValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

func (sv scriptValue) marshal() interface{} {
	switch cv := sv.LValue.(type) {
	case lua.LBool, lua.LNumber, lua.LString:
		return cv
	case *lua.LTable:
		if sv.visited[cv] {
			panic(errScriptMarshalNested)
		}
		sv.visited[cv] = true

		var arr []interface{}
		var obj map[string]interface{}
		cv.ForEach(func(k lua.LValue, v lua.LValue) {
			i, numberKey := k.(lua.LNumber)
			if numberKey && obj == nil {
				index := int(i) - 1
				if index != len(arr) {
					// map out of order; convert to map
					obj = make(map[string]interface{})
					for i, value := range arr {
						obj[strconv.Itoa(i+1)] = value
					}
					obj[strconv.Itoa(index+1)] = scriptValue{v, sv.visited}.marshal()
					return
				}
				arr = append(arr, scriptValue{v, sv.visited}.marshal())
				return
			}
			if obj == nil {
				obj = make(map[string]interface{})
				for i, value := range arr {
					obj[strconv.Itoa(i+1)] = value
				}
			}
			obj[k.String()] = scriptValue{v, sv.visited}.marshal()
		})
		if obj != nil {
			return obj
		}
		return arr
	case *lua.LFunction:
		panic(errScriptMarshalFunction)
	case lua.LChannel:
		panic(errScriptMarshalChannel)
	case *lua.LState:
		panic(errScriptMarshalState)
	case *lua.LUserData:
		panic(errScriptMarshalUserData)
	}
	return nil
}

func scriptValueMarshal(v lua.LValue) interface{} {
	return scriptValue{v, make(map[*lua.LTable]bool)}.marshal()
}
