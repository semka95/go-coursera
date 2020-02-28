package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const filePath string = "./data/dataset.xml"

var (
	ts = httptest.NewServer(http.HandlerFunc(SearchServer))
)

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
	case "__broken_error_json":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status": 400`)
		return
	case "__internal_error":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "__timeout_error":
		time.Sleep(2 * time.Second)
		return
	case "__bad_request":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error": "test bad request"}`)
		return
	case "__broken_result_json":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"name": "test`)
		return
	}

	if !(orderField == "Id" || orderField == "Age" || orderField == "Name") {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error": "ErrorBadOrderField"}`)
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
}

func getRecords(query string) []User {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	users := make([]User, 100)
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
			switch tok.Name.Local {
			case "row":
				u = User{}
			case "id":
				temp := 0
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				u.Id = temp
			case "first_name":
				temp := ""
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				firstName = temp
			case "last_name":
				temp := ""
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				lastName = temp
			case "age":
				temp := 0
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				u.Age = temp
			case "about":
				temp := ""
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				u.About = temp
			case "gender":
				temp := ""
				if err := decoder.DecodeElement(&temp, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				u.Gender = temp
			}
		case xml.EndElement:
			if tok.Name.Local == "row" {
				u.Name = firstName + " " + lastName
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
		Error  string
	}{
		{
			Name: "test limit less then 0",
			Input: SearchRequest{
				Limit: -1,
			},
			Error: "limit must be > 0",
		},
		{
			Name: "test offset less then 0",
			Input: SearchRequest{
				Offset: -1,
			},
			Error: "offset must be > 0",
		},
		{
			Name: "test server internal error",
			Input: SearchRequest{
				Query: "__internal_error",
			},
			Error: "SearchServer fatal error",
		},
		{
			Name: "test broken error json",
			Input: SearchRequest{
				Query: "__broken_error_json",
			},
			Error: "cant unpack error json: unexpected end of JSON input",
		},
		{
			Name: "test bad order field",
			Input: SearchRequest{
				OrderField: "wrong",
				Limit:      5,
				Offset:     1,
				Query:      "test",
			},
			Error: "OrderFeld wrong invalid",
		},
		{
			Name: "test bad request and limit higher than 25",
			Input: SearchRequest{
				Query:      "__bad_request",
				OrderField: "Id",
				Limit:      26,
				Offset:     1,
			},
			Error: "unknown bad request error: test bad request",
		},
		{
			Name: "test broken result json",
			Input: SearchRequest{
				Query:      "__broken_result_json",
				OrderField: "Id",
				Limit:      10,
				Offset:     1,
			},
			Error: "cant unpack result json: unexpected end of JSON input",
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			s := &SearchClient{
				AccessToken: "ok",
				URL:         ts.URL,
			}

			_, err := s.FindUsers(test.Input)
			assertError(t, err, test.Error)
		})
	}
}

func TestFindUsers__Timeout(t *testing.T) {
	req := SearchRequest{
		Query: "__timeout_error",
	}
	searcherParams := newSearcherParams(&req)

	want := fmt.Sprintf("timeout for %s", searcherParams.Encode())

	s := &SearchClient{
		AccessToken: "ok",
		URL:         ts.URL,
	}

	_, err := s.FindUsers(req)
	assertError(t, err, want)
}

func TestFindUsers__IncorrectURL(t *testing.T) {
	req := SearchRequest{}
	searcherParams := newSearcherParams(&req)
	s := &SearchClient{
		AccessToken: "ok",
		URL:         "http://127.0.0.1:123",
	}

	want := fmt.Sprintf("unknown error Get %s?%s: dial tcp %s: connect: connection refused", s.URL, searcherParams.Encode(), s.URL[7:])

	_, err := s.FindUsers(req)

	assertError(t, err, want)
}

func TestFindUsers__BadAccessToken(t *testing.T) {
	req := SearchRequest{}

	s := &SearchClient{
		AccessToken: "bad_token",
		URL:         ts.URL,
	}

	want := "Bad AccessToken"

	_, err := s.FindUsers(req)

	assertError(t, err, want)
}

func TestFindUsers__Ok(t *testing.T) {
	req := SearchRequest{
		Query:      "esse",
		OrderField: "Id",
		Limit:      10,
		Offset:     1,
		OrderBy:    0,
	}

	newSearcherParams(&req)

	s := &SearchClient{
		AccessToken: "ok",
		URL:         ts.URL,
	}

	want := ""

	_, err := s.FindUsers(req)

	assertError(t, err, want)
}

func newSearcherParams(req *SearchRequest) url.Values {
	searcherParams := url.Values{}
	searcherParams.Add("limit", strconv.Itoa(req.Limit+1))
	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	searcherParams.Add("query", req.Query)
	searcherParams.Add("order_field", req.OrderField)
	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))
	return searcherParams
}

func assertError(t *testing.T, got error, want string) {
	t.Helper()
	if got == nil {
		t.Fatal("didn't get an error but wanted one")
	}

	if got.Error() != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
