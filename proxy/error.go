package proxy

import (
	"fmt"
	"html/template"
	"net/http"
)

type proxyError struct {
	status int

	String string
	Route  *route
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
