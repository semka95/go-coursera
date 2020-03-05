package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		srv.handlerProfile(w, r)
	case "/user/create":
		srv.handlerCreate(w, r)
	default:
		http.Error(w, `{"error": "unknown method"}`, http.StatusNotFound)
	}
}

func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fill params struct
	params := ProfileParams{
		Login: r.FormValue("login"),
	}

	// validate params
	if params.Login == "" {
		http.Error(w, `{"error": "login must me not empty"}`, http.StatusBadRequest)
		return
	}

	// get data from method
	u, err := srv.Profile(context.Background(), params)
	// check errors
	if err != nil {
		var e ApiError
		errText := fmt.Sprintf(`{"error": "%s"}`, err)
		if errors.As(err, &e) {
			http.Error(w, errText, e.HTTPStatus)
			return
		}
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	// prepare structure to write to response
	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	// write to response
	answer, _ := json.Marshal(result)
	w.Write(answer)
}

func (srv *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "bad method"}`, http.StatusNotAcceptable)
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		return
	}

	age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		http.Error(w, `{"error": "age must be int"}`, http.StatusBadRequest)
		return
	}

	params := CreateParams{
		Login:  r.FormValue("login"),
		Name:   r.FormValue("full_name"),
		Status: r.FormValue("status"),
		Age:    age,
	}

	// validation
	if params.Login == "" {
		http.Error(w, `{"error": "login must me not empty"}`, http.StatusBadRequest)
		return
	}

	if len(params.Login) < 10 {
		http.Error(w, `{"error": "login len must be >= 10"}`, http.StatusBadRequest)
		return
	}

	if params.Status == "" {
		params.Status = "user"
	}

	if !(params.Status == "user" || params.Status == "moderator" || params.Status == "admin") {
		http.Error(w, `{"error": "status must be one of [user, moderator, admin]"}`, http.StatusBadRequest)
		return
	}

	if params.Age < 0 {
		http.Error(w, `{"error": "age must be >= 0"}`, http.StatusBadRequest)
		return
	}

	if params.Age > 128 {
		http.Error(w, `{"error": "age must be <= 128"}`, http.StatusBadRequest)
		return
	}

	u, errA := srv.Create(context.Background(), params)
	// check errors
	if errA != nil {
		var e ApiError
		errText := fmt.Sprintf(`{"error": "%s"}`, errA)
		if errors.As(errA, &e) {
			http.Error(w, errText, e.HTTPStatus)
			return
		}
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	// prepare structure to write to response
	// resp, _ := json.Marshal(u)
	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	// write to response
	answer, _ := json.Marshal(result)
	w.Write(answer)

}
