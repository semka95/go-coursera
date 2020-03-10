package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
			fmt.Println(err.Error())
			return
		}
	} else {
		path := strings.Split(r.URL.Path, "/")[1:]
		if len(path) == 1 {
			err := h.getTableRecords(path[0], r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
			if err != nil {
				w.WriteHeader(h.Error.Code)
				w.Write([]byte(h.Error.Message))
				fmt.Println(err.Error())
				return
			}
		}
		if len(path) == 2 {
			err := h.getTableRecord(path[0], path[1])
			if err != nil {
				w.WriteHeader(h.Error.Code)
				w.Write([]byte(h.Error.Message))
				fmt.Println(err.Error())
				return
			}
		}

	}

	answer, err := json.Marshal(h.Result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshaling answer"}`))
	}
	w.Write(answer)
}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	table := strings.ReplaceAll(r.URL.Path, "/", "")
	decoder := json.NewDecoder(r.Body)
	var postVals map[string]interface{}
	err := decoder.Decode(&postVals)
	idField := h.getPKColumnName(table)

	colTypes, err := h.getTableColumns(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		fmt.Println(err.Error())
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
				case "text":
					postVals[k] = new(string)
				case "varchar":
					postVals[k] = new(string)
				case "float":
					postVals[k] = new(float64)
				}
			}
		}
	}
	delete(postVals, idField)

	var query strings.Builder
	var query2 strings.Builder
	fmt.Fprintf(&query, "INSERT INTO %v (", table)
	fmt.Fprintf(&query2, ") VALUES (")
	values := make([]interface{}, 0, len(postVals))
	for k, v := range postVals {
		values = append(values, v)
		fmt.Fprintf(&query, "`%s`", k)
		fmt.Fprintf(&query2, "?")
		if len(values) != len(postVals) {
			fmt.Fprintf(&query, ", ")
			fmt.Fprintf(&query2, ", ")
		}
	}

	fmt.Fprintf(&query2, ")")
	resQuery := query.String() + query2.String()
	fmt.Println(resQuery)
	result, err := h.DB.Exec(resQuery, values...)
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		fmt.Println(err.Error())
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		fmt.Println(err.Error())
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
	}
	w.Write(answer)
}

func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")[1:]
	table := path[0]
	id := path[1]
	idField := h.getPKColumnName(table)

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	var postVals map[string]interface{}
	err := decoder.Decode(&postVals)

	colTypes, err := h.getTableColumns(table)
	if err != nil {
		w.WriteHeader(h.Error.Code)
		w.Write([]byte(h.Error.Message))
		fmt.Println(err.Error())
		return
	}

	for k, v := range postVals {
		// if _, ok := colTypes[k]; !ok {
		// 	delete(postVals, k)
		// 	continue
		// }

		if v == nil {
			if !colTypes[k].IsNullable {
				w.WriteHeader(http.StatusBadRequest)
				e := fmt.Sprintf(`{"error": "field %s have invalid type"}`, k)
				w.Write([]byte(e))
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

		if colTypes[k].Type == "text" || colTypes[k].Type == "varchar" {
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
			return
		}

	}

	var query strings.Builder
	fmt.Fprintf(&query, "UPDATE %v SET ", table)
	values := make([]interface{}, 0, len(postVals))
	i := 0
	for k, v := range postVals {
		if v == nil {
			i++
			fmt.Fprintf(&query, "`%s` = NULL", k)
			if i != len(postVals) {
				fmt.Fprintf(&query, ", ")
			}
			continue
		}
		values = append(values, v)
		i++
		fmt.Fprintf(&query, "`%s` = ?", k)
		if i != len(postVals) {
			fmt.Fprintf(&query, ", ")
		}
	}

	fmt.Fprintf(&query, " WHERE %s = ?", idField)
	resQuery := query.String()
	values = append(values, id)

	result, err := h.DB.Exec(resQuery, values...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "internal server error"`))
		fmt.Println(err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		fmt.Println(err.Error())
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
	}
	w.Write(answer)

}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")[1:]
	table := path[0]
	id := path[1]
	idField := h.getPKColumnName(table)

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, idField)
	result, err := h.DB.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "internal server error"`))
		fmt.Println(err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		fmt.Println(err.Error())
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
		table := ""
		err = rows.Scan(&table)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `{"error": "internal server error"}`
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
	limit, err := strconv.Atoi(l)
	if err != nil {
		limit = 5
	}

	offset, err := strconv.Atoi(o)
	if err != nil {
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
		value := make([]interface{}, len(columns))
		for i := range columns {
			switch st := columns[i].DatabaseTypeName(); st {
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

	rows.Close()

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

	idField := h.getPKColumnName(table)

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
	value := make([]interface{}, len(columns))
	for i := range columns {
		st := columns[i].DatabaseTypeName()
		switch st {
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

	rows.Close()

	h.Result = map[string]interface{}{
		"response": map[string]interface{}{
			"record": item,
		},
	}

	return nil
}

func (h *Handler) getPKColumnName(table string) string {
	query := fmt.Sprintf("SELECT `COLUMN_NAME` FROM `information_schema`.`COLUMNS` WHERE (`TABLE_NAME` = '%s') AND (`COLUMN_KEY` = 'PRI')", table)
	row := h.DB.QueryRow(query)
	idField := new(string)
	row.Scan(&idField)
	return *idField
}

type column struct {
	Type       string
	IsNullable bool
}

func (h *Handler) getTableColumns(table string) (map[string]column, error) {
	str1 := ""
	str2 := ""
	isNullstr := ""
	result := make(map[string]column)

	query := fmt.Sprintf("SELECT `COLUMN_NAME`, `DATA_TYPE`, `IS_NULLABLE` FROM `information_schema`.`COLUMNS` WHERE (`TABLE_NAME` = '%s');", table)

	rows, err := h.DB.Query(query)
	if err != nil {
		h.Error.Code = http.StatusInternalServerError
		h.Error.Message = `{"error": "internal server error"}`
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&str1, &str2, &isNullstr)
		if err != nil {
			h.Error.Code = http.StatusInternalServerError
			h.Error.Message = `{"error": "internal server error"}`
			return nil, err
		}

		var isNull bool
		if isNullstr == "NO" {
			isNull = false
		} else {
			isNull = true
		}

		result[str1] = column{Type: str2, IsNullable: isNull}
	}

	rows.Close()
	return result, nil
}
