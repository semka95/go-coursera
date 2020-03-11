package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "unknown method"}`))
	}
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		err := h.listTables()
		if err != nil {
			w.WriteHeader(h.Error.Code)
			w.Write([]byte(h.Error.Message))
			log.Println(err.Error())
		}
	} else {
		path := strings.Split(r.URL.Path, "/")[1:]
		if len(path) == 1 {
			err := h.getTableRecords(path[0], r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
			if err != nil {
				w.WriteHeader(h.Error.Code)
				w.Write([]byte(h.Error.Message))
				log.Println(err.Error())
				return
			}
		}
		if len(path) == 2 {
			err := h.getTableRecord(path[0], path[1])
			if err != nil {
				w.WriteHeader(h.Error.Code)
				w.Write([]byte(h.Error.Message))
				log.Println(err.Error())
				return
			}
		}
	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshaling answer"}`))
		return
	}
	w.Write(answer)
}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	table := strings.ReplaceAll(r.URL.Path, "/", "")

	decoder := json.NewDecoder(r.Body)
	var postVals map[string]interface{}
	err := decoder.Decode(&postVals)

	idField, err := h.getPKColumnName(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		log.Println(err.Error())
		return
	}

	colTypes, err := h.getTableColumns(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		log.Println(err.Error())
		return
	}

	for k := range postVals {
		if _, ok := colTypes[k]; !ok {
			delete(postVals, k)
		}
	}

	for k, v := range colTypes {
		if !v.IsNullable {
			if _, ok := postVals[k]; !ok {
				switch v.Type {
				case "int":
					postVals[k] = new(int64)
				case "string":
					postVals[k] = new(string)
				case "float":
					postVals[k] = new(float64)
				}
			}
		}
	}
	delete(postVals, idField)

	var fields strings.Builder
	var placeholders strings.Builder
	values := make([]interface{}, 0, len(postVals))
	for k, v := range postVals {
		values = append(values, v)
		fmt.Fprintf(&fields, "`%s`", k)
		fmt.Fprintf(&placeholders, "?")
		if len(values) != len(postVals) {
			fmt.Fprintf(&fields, ", ")
			fmt.Fprintf(&placeholders, ", ")
		}
	}

	resQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, fields.String(), placeholders.String())
	result, err := h.DB.Exec(resQuery, values...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		log.Println(err.Error())
		return
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		log.Println(err.Error())
		return
	}

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			idField: lastID,
		},
	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshaling answer"}`))
		log.Println(err.Error())
		return
	}
	w.Write(answer)
}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")[1:]
	table := path[0]
	id := path[1]
	idField, err := h.getPKColumnName(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		log.Println(err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	var postVals map[string]interface{}
	err = decoder.Decode(&postVals)

	colTypes, err := h.getTableColumns(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		log.Println(err.Error())
		return
	}

	for k, v := range postVals {
		if _, ok := colTypes[k]; !ok {
			delete(postVals, k)
			continue
		}

		if v == nil {
			if !colTypes[k].IsNullable {
				w.WriteHeader(http.StatusBadRequest)
				e := fmt.Sprintf(`{"error": "field %s have invalid type"}`, k)
				w.Write([]byte(e))
				log.Printf("tried to assign null value to not nullable field %s\n", k)
				return
			}
		}

		valid := false
		if colTypes[k].Type == "int" {
			val, err := v.(json.Number).Int64()
			if err != nil {
				valid = false
			} else {
				valid = true
				postVals[k] = val
			}
		}

		if colTypes[k].Type == "string" {
			if _, ok := v.(json.Number); ok {
				valid = false
			} else {
				valid = true
			}
		}

		if colTypes[k].Type == "float" {
			val, err := v.(json.Number).Float64()
			if err != nil {
				valid = false
			} else {
				valid = true
				postVals[k] = val
			}
		}

		if k == idField {
			valid = false
		}

		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			e := fmt.Sprintf(`{"error": "field %s have invalid type"}`, k)
			w.Write([]byte(e))
			log.Printf("field %s have invalid type", k)
			return
		}
	}

	var query strings.Builder
	values := make([]interface{}, 0, len(postVals))
	for k, v := range postVals {
		values = append(values, v)
		fmt.Fprintf(&query, "`%s` = ?", k)
		if len(values) != len(postVals) {
			fmt.Fprintf(&query, ", ")
		}
	}

	resQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", table, query.String(), idField)
	values = append(values, id)

	result, err := h.DB.Exec(resQuery, values...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "internal server error"`))
		log.Println(err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		log.Println(err.Error())
		return
	}

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			"updated": affected,
		},
	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshaling answer"}`))
		log.Println(err.Error())
		return
	}
	w.Write(answer)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")[1:]
	table := path[0]
	id := path[1]

	idField, err := h.getPKColumnName(table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "internal server error"`))
		log.Println(err.Error())
		return
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, idField)
	result, err := h.DB.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "internal server error"`))
		log.Println(err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		log.Println(err.Error())
		return
	}

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			"deleted": affected,
		},
	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshaling answer"}`))
		log.Println(err.Error())
		return
	}
	w.Write(answer)
}

func (h *Handler) listTables() error {
	var items []string

	rows, err := h.DB.Query("SHOW TABLES")
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `{"error": "internal server error"}`
			return err
		}
		items = append(items, table)
	}

	if err = rows.Err(); err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	err = rows.Close()
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	tables := map[string][]string{"tables": items}

	h.Result = map[string]interface{}{
		"response": tables,
	}

	return nil
}

func (h *Handler) getTableRecords(table, l, o string) error {
	limit, err := strconv.Atoi(l)
	if err != nil || limit <= 0 {
		limit = 5
	}

	offset, err := strconv.Atoi(o)
	if err != nil || offset < 0 {
		offset = 0
	}

	var items []map[string]interface{}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
	rows, err := h.DB.Query(query, limit, offset)
	if err != nil {
		me, _ := err.(*mysql.MySQLError)
		if me.Number == 1146 {
			h.Error.Code = http.StatusNotFound
			h.Error.Message = `{"error": "unknown table"}`
			return err
		}
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}
	columns, _ := rows.ColumnTypes()

	for rows.Next() {
		value := prepareScanValue(columns)
		err = rows.Scan(value...)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `{"error": "internal server error"}`
			return err
		}

		item := make(map[string]interface{})
		for i := range value {
			item[columns[i].Name()] = value[i]
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	err = rows.Close()
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			"records": items,
		},
	}

	return nil
}

func (h *Handler) getTableRecord(table, id string) error {
	rec, err := strconv.Atoi(id)
	if err != nil {
		h.Error.Code = http.StatusBadRequest
		h.Error.Message = `{"error": "bad id"}`
		return err
	}

	idField, err := h.getPKColumnName(table)
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, idField)
	rows, err := h.DB.Query(query, rec)
	if err != nil {
		me, _ := err.(*mysql.MySQLError)
		if me.Number == 1146 {
			h.Error.Code = http.StatusNotFound
			h.Error.Message = `{"error": "unknown table"}`
			return err
		}
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}
	columns, _ := rows.ColumnTypes()

	rows.Next()
	value := prepareScanValue(columns)
	err = rows.Scan(value...)
	if err != nil {
		h.Error.Code = http.StatusNotFound
		h.Error.Message = `{"error": "record not found"}`
		return err
	}

	item := make(map[string]interface{})
	for i := range value {
		item[columns[i].Name()] = value[i]
	}

	if err = rows.Err(); err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	err = rows.Close()
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return err
	}

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			"record": item,
		},
	}

	return nil
}

func (h *Handler) getPKColumnName(table string) (string, error) {
	query := fmt.Sprintf("SELECT `COLUMN_NAME` FROM `information_schema`.`COLUMNS` WHERE (`TABLE_NAME` = '%s') AND (`COLUMN_KEY` = 'PRI')", table)
	row := h.DB.QueryRow(query)
	idField := new(string)
	err := row.Scan(&idField)
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return "", err
	}
	return *idField, nil
}

type column struct {
	Type       string
	IsNullable bool
}

type columnsInfo map[string]column

func (h *Handler) getTableColumns(table string) (columnsInfo, error) {
	var colName, colType, isNulls string
	result := make(columnsInfo)

	query := fmt.Sprintf("SELECT `COLUMN_NAME`, `DATA_TYPE`, `IS_NULLABLE` FROM `information_schema`.`COLUMNS` WHERE (`TABLE_NAME` = '%s');", table)

	rows, err := h.DB.Query(query)
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&colName, &colType, &isNulls)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `{"error": "internal server error"}`
			return nil, err
		}

		isNull := true
		if isNulls == "NO" {
			isNull = false
		}

		if colType == "text" || colType == "varchar" {
			colType = "string"
		}

		result[colName] = column{Type: colType, IsNullable: isNull}
	}

	if err = rows.Err(); err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return nil, err
	}

	err = rows.Close()
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return nil, err
	}
	return result, nil
}

func prepareScanValue(columns []*sql.ColumnType) []interface{} {
	value := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i].DatabaseTypeName() {
		case "INT":
			value[i] = new(*int64)
		case "FLOAT":
			value[i] = new(*float64)
		case "VARCHAR":
			value[i] = new(*string)
		case "TEXT":
			value[i] = new(*string)
		}
	}
	return value
}
