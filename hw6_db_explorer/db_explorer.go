package main

import (
	"database/sql"
	"net/http"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type Handler struct {
	DB *sql.DB
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	mux := http.NewServeMux()
	handler := &Handler{DB: db}
	mux.Handle("/", handler)
	return mux, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, `{"error": "unknown method"}`, http.StatusNotFound)
	}

}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {

}
