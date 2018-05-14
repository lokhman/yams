package proxy

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/lokhman/yams/proxy/model"
	"github.com/lokhman/yams/yams"
)

const (
	skipPanic = 4
	skipError = 1
)

type proxyError struct {
	status int
	String string
	Route  *model.Route
	Caller string
}

func perror(w http.ResponseWriter, status int, str string, r *model.Route, skip int) {
	var caller string
	if yams.Debug {
		if pc, file, line, ok := runtime.Caller(skip); ok {
			caller = fmt.Sprintf("%s:%d (0x%x)", file, line, pc)
		}
	}
	proxyError{status, str, r, caller}.write(w)
}

func (pe proxyError) write(w http.ResponseWriter) {
	if pe.Route != nil && !pe.Route.Profile.Debug {
		http.Error(w, fmt.Sprintf("%d %s", pe.status, http.StatusText(pe.status)), pe.status)
		return
	}

	w.Header().Set(yams.ProxyHeaderStatus, yams.ProxyStatusError)
	w.WriteHeader(pe.status)

	t, err := template.ParseFiles("public/error.html")
	if err != nil {
		panic(err)
	}

	if err = t.Execute(w, pe); err != nil {
		panic(err)
	}
}
