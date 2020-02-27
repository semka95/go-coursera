package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const filePath string = "daraset.xml"

// код писать тут
func SearchServer(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	if orderField == "" {
		orderField = "Name"
	}

	accessToken := r.Header.Get("AccessToken")
	if accessToken == "bad_token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch query {
	case "__broken_json":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status": 400`)
		return
	case "__internal_error":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "__timeout_error":
		time.Sleep(2 * time.Second)
		return
	}

	if !(orderField == "Id" || orderField == "Age" || orderField == "Name") {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error": "OrderField invalid"`)
		return
	}

	// add errors
	records := getRecords(query)
	/// !!! check
	records = records[offset : offset+limit-1]
	sortRecords(records, orderField)

	json, err := json.Marshal(records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	w.WriteHeader(http.StatusOK)
}

func getRecords(query string) []User {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	users := make([]User, 0)
	u := User{}
	firstName := ""
	lastName := ""

	decoder := xml.NewDecoder(file)

	for {
		tok, tokenErr := decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			fmt.Println("error happend", tokenErr)
			break
		} else if tokenErr == io.EOF {
			break
		}
		if tok == nil {
			fmt.Println("t is nil break")
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			var temp interface{}
			if err := decoder.DecodeElement(temp, &tok); err != nil {
				fmt.Println("error happend", err)
			}
			switch tok.Name.Local {
			case "row":
				u = User{}
			case "id":
				u.Id = temp.(int)
			case "first_name":
				firstName = temp.(string)
			case "last_name":
				lastName = temp.(string)
			case "age":
				u.Age = temp.(int)
			case "about":
				u.About = temp.(string)
			case "gender":
				u.Gender = temp.(string)
			}
		case xml.EndElement:
			if tok.Name.Local == "row" {
				u.Name = firstName + lastName
				if strings.Contains(u.About, query) || strings.Contains(u.Name, query) {
					users = append(users, u)
				}
			}
		}
	}

	return users
}

func fillUser(decoder *xml.Decoder, tok *xml.StartElement, u *User) {

}

func sortRecords(users []User, orderField string) {
	switch orderField {
	case "Name":
		// !!! check
		sort.Slice(users, func(i, j int) bool { return strings.Compare(users[i].Name, users[j].Name) == -1 })
	case "Age":
		sort.Slice(users, func(i, j int) bool { return users[i].Age < users[j].Age })
	case "Id":
		sort.Slice(users, func(i, j int) bool { return users[i].Id < users[j].Id })
	}
}

func TestFindUsers(t *testing.T) {
	cases := []struct {
		Name   string
		Input  SearchRequest
		Output SearchResponse
	}{
		{
			Name: "test limit < 0",
			Input: SearchRequest{
				Limit:      -1,
				Offset:     1,
				Query:      "test",
				OrderField: "Id",
				OrderBy:    -1,
			},
			Output: SearchResponse{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			s := &SearchClient{
				AccessToken: "ok",
				URL:         ts.URL,
			}

			result, err := s.FindUsers(test.Input)
		})
	}
}
