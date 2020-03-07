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

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		srv.handlerCreate(w, r)
	default:
		http.Error(w, `{"error": "unknown method"}`, http.StatusNotFound)
	}
}

func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := ProfileParams{
		Login: r.FormValue("login"),
	}

	if params.Login == "" {
		http.Error(w, `{"error": "login must me not empty"}`, http.StatusBadRequest)
		return
	}

	u, err := srv.Profile(context.Background(), params)
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

	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	answer, err := json.Marshal(result)
	if err != nil {
		http.Error(w, `{"error": "error marshaling answer"}`, http.StatusInternalServerError)
	}
	w.Write(answer)
}

func (srv *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
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

	u, err := srv.Create(context.Background(), params)
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

	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	answer, err := json.Marshal(result)
	if err != nil {
		http.Error(w, `{"error": "error marshaling answer"}`, http.StatusInternalServerError)
	}
	w.Write(answer)
}

func (srv *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"error": "bad method"}`, http.StatusNotAcceptable)
		return
	}
	if r.Header.Get("X-Auth") != "100500" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		return
	}

	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		http.Error(w, `{"error": "level must be int"}`, http.StatusBadRequest)
		return
	}

	params := OtherCreateParams{
		Username: r.FormValue("username"),
		Name:     r.FormValue("account_name"),
		Class:    r.FormValue("class"),
		Level:    level,
	}

	if params.Username == "" {
		http.Error(w, `{"error": "username must me not empty"}`, http.StatusBadRequest)
		return
	}

	if len(params.Username) < 3 {
		http.Error(w, `{"error": "username len must be >= 3"}`, http.StatusBadRequest)
		return
	}

	if params.Class == "" {
		params.Class = "warrior"
	}

	if !(params.Class == "warrior" || params.Class == "sorcerer" || params.Class == "rouge") {
		http.Error(w, `{"error": "class must be one of [warrior, sorcerer, rouge]"}`, http.StatusBadRequest)
		return
	}

	if params.Level < 1 {
		http.Error(w, `{"error": "level must be >= 1"}`, http.StatusBadRequest)
		return
	}

	if params.Level > 50 {
		http.Error(w, `{"error": "level must be <= 50"}`, http.StatusBadRequest)
		return
	}

	u, err := srv.Create(context.Background(), params)
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

	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	answer, err := json.Marshal(result)
	if err != nil {
		http.Error(w, `{"error": "error marshaling answer"}`, http.StatusInternalServerError)
	}
	w.Write(answer)
}
