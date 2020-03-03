package main

import (
	"net/http"
)

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		srv.handlerProfile(w, r)
	case "/user/create":
		srv.handlerCreate(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	params := ProfileParams{
		Login: r.URL.Query().Get("login"),
	}

	if params.Login == "" {
		http.Error(w, `{"error": "login must me not empty"}`, http.StatusBadRequest)
		return
	}

}
