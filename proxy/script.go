package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/lokhman/yams-lua"
	"github.com/lokhman/yams-lua/json"
)

type script struct {
	route *route
	resp  http.ResponseWriter
	req   *http.Request
}

func (s *script) execute() error {
	l := lua.NewState()
	defer l.Close()

	json.Preload(l)
	l.PreloadModule("yams", s.loader)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.route.timeout)*time.Second)
	defer cancel()

	l.SetContext(ctx)
	return l.DoString(s.route.script)
}

func (s *script) loader(l *lua.LState) int {
	mod := l.NewTable()

	// constants
	l.SetField(mod, "id", lua.LString(s.route.uuid))
	l.SetField(mod, "method", lua.LString(s.req.Method))
	l.SetField(mod, "host", lua.LString(s.req.Host))
	l.SetField(mod, "uri", lua.LString(s.req.URL.Path))
	l.SetField(mod, "querystring", lua.LString(s.req.URL.RawQuery))
	l.SetField(mod, "remoteaddr", lua.LString(s.req.RemoteAddr))

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
		"getheader": s.fnGetHeader,
		"setheader": s.fnSetHeader,
		"setcode":   s.fnSetCode,
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

func (s *script) fnGetHeader(l *lua.LState) int {
	key := l.CheckString(1)
	value := s.req.Header.Get(key)
	l.Push(lua.LString(value))
	return 1
}

func (s *script) fnSetHeader(l *lua.LState) int {
	key := l.CheckString(1)
	value := l.CheckString(2)
	s.resp.Header().Set(key, value)
	for i := 3; i <= l.GetTop(); i++ {
		value = l.CheckString(i)
		s.resp.Header().Add(key, value)
	}
	return 0
}

func (s *script) fnSetCode(l *lua.LState) int {
	s.resp.WriteHeader(int(l.CheckNumber(1)))
	return 0
}

func (s *script) fnGet(l *lua.LState) int {
	key := l.CheckString(1)
	if v, ok := s.req.URL.Query()[key]; ok && len(v) > 0 {
		l.Push(lua.LString(v[0]))
	} else if v, ok := s.route.params[key]; ok {
		l.Push(lua.LString(v))
	} else if v, ok := s.req.PostForm[key]; ok && len(v) > 0 {
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
				assetLoad(s.resp, v)
				continue
			}
		}
		fmt.Fprint(s.resp, v)
	}
	return 0
}

func (s *script) fnGetVar(l *lua.LState) int {
	return 0
}

func (s *script) fnSetVar(l *lua.LState) int {
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
	s.route.profile.proxy(s.resp, s.req)
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
	l.Push(lua.LString(assetRead(assetCheck(l))))
	return 1
}

func assetFnTemplate(l *lua.LState) int {
	a := assetCheck(l)
	t := template.Must(template.New(a.path).Parse(string(assetRead(a))))
	v := scriptValue{l.CheckTable(2), make(map[*lua.LTable]bool)}.marshal()
	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, v); err != nil {
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
