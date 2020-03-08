package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	URL   string `xml:"guid"`
	Title string `xml:"title"`
}

func GetHabrPosts() (*RSS, error) {
	fmt.Println("fetching https://habr.com/ru/rss/best/")
	resp, err := http.Get("https://habr.com/ru/rss/best/")
	if err != nil {
		fmt.Println("get error")
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	rss := new(RSS)
	err = xml.Unmarshal(body, rss)
	if err != nil {
		fmt.Println("unmarshal error")
		return nil, err
	}

	return rss, nil
}
