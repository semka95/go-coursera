package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type HandlerError struct {
	Code    int
	Message string
}

type Handler struct {
	DB     *sql.DB
	Result map[string]interface{}
	Error  HandlerError
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	mux := http.NewServeMux()
	handler := &Handler{DB: db}
	mux.Handle("/", handler)
	return mux, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	if r.URL.Path == "/" {
		err := h.listTables()
		if err != nil {
			http.Error(w, h.Error.Message, h.Error.Code)
			fmt.Println(err.Error())
			return
		}
	} else {
		path := strings.Split(r.URL.Path, "/")[1:]
		if len(path) == 1 {
			err := h.getTableRecords(path[0], r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		}

	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		http.Error(w, `{"error": "error marshaling answer"}`, http.StatusInternalServerError)
	}
	w.Write(answer)
}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) listTables() error {
	var items []string

	rows, err := h.DB.Query("SHOW TABLES")
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `"error": "internal server error"`
		return err
	}

	for rows.Next() {
		table := ""
		err = rows.Scan(&table)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `"error": "internal server error"`
			return err
		}
		items = append(items, table)
	}

	rows.Close()

	tables := map[string][]string{"tables": items}

	h.Result = map[string]interface{}{
		"response": tables,
	}

	return nil
}

func (h *Handler) getTableRecords(table, l, o string) error {
	limit := 5
	offset := 0
	if l != "" {
		limit, err := strconv.Atoi(l)
	}
	return nil
}
