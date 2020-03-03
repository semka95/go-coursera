package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	client = http.Client{Timeout: time.Duration(time.Millisecond)}
)

func getRemoteResource() error {
	url := "http://127.0.0.1:9999/pages?id=123"
	_, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("resource error: %w", err)
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := getRemoteResource()
	if err != nil {
		fmt.Printf("full err: %+v\n", err)
		var e *url.Error
		if errors.As(err, &e) {
			fmt.Printf("resource %s err: %+v\n", e.URL, e.Err)
			http.Error(w, "remote resource error", 500)
			return
		}
		fmt.Printf("%+v\n", err)
		http.Error(w, "parsing error", 500)
		return
	}
	w.Write([]byte("all is OK"))
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
