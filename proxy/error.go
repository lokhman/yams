package proxy

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/lokhman/yams/yams"
)

type proxyError struct {
	status int
	String string
	Route  *route
	Caller string
}

func perr(w http.ResponseWriter, status int, str string, r *route) {
	var caller string
	if yams.IsDebug() {
		if pc, file, line, ok := runtime.Caller(4); ok {
			caller = fmt.Sprintf("%s:%d (0x%x)", file, line, pc)
		}
	}
	proxyError{status, str, r, caller}.write(w)
}

func (pe proxyError) write(w http.ResponseWriter) {
	if pe.Route != nil && !pe.Route.profile.debug {
		http.Error(w, fmt.Sprintf("%d %s", pe.status, http.StatusText(pe.status)), pe.status)
		return
	}

	w.Header().Set(headerStatus, statusError)
	w.WriteHeader(pe.status)

	t, err := template.ParseFiles("public/error.html")
	if err != nil {
		panic(err)
	}

	if err = t.Execute(w, pe); err != nil {
		panic(err)
	}
}
