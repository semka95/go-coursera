package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(file)

	seenBrowsers := make(map[uint32]struct{})
	var foundUsers strings.Builder
	userNum := 0

	for {
		isAndroid := false
		isMSIE := false

		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		switch t.(type) {
		case json.Delim:
			dec.More()
		case string:
			tok, ok := t.(string)
			if !ok {
				log.Println("cant cast token to string 41")
				continue
			}
			if tok == "browsers" {
				handleBrowser(&isAndroid, &isMSIE, dec, seenBrowsers)
				if !(isAndroid && isMSIE) {
					continue // check for switch
				}
				handleNameEmail(dec, &foundUsers, userNum)
				userNum++
			}
			dec.More()
		}

	}

	fmt.Fprintln(out, "found users:\n"+foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func handleNameEmail(dec *json.Decoder, foundUsers *strings.Builder, userNum int) {
	name := ""
	email := ""
	dec.More() //?
	for b := dec.More(); b; {
		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		tok, ok := t.(string) // refactor
		if !ok {
			log.Println("cant cast token to string")
			continue
		}

		switch tok {
		case "email":
			dec.More()
			t, err := dec.Token()
			if checkTokenErr(t, err) {
				break // this is not working here
			}

			email, ok = t.(string) // refactor
			if !ok {
				log.Println("cant cast token to string")
				continue
			}
			email = strings.ReplaceAll(tok, "@", " [at] ")

		case "name":
			dec.More()
			t, err := dec.Token()
			if checkTokenErr(t, err) {
				break // this is not working here
			}

			name, ok = t.(string) // refactor
			if !ok {
				log.Println("cant cast token to string")
				continue
			}
		}

	}
	fmt.Fprintf(foundUsers, "[%d] %s <%s>\n", userNum, name, email)
}

func handleBrowser(isAndroid, isMSIE *bool, dec *json.Decoder, seenBrowsers map[uint32]struct{}) {
	dec.More()
	for {
		dec.More()
		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		if _, ok := t.(json.Delim); ok {
			break
		}

		tok, ok := t.(string) // refactor
		if !ok {
			log.Println("cant cast token to string 120", t)
			continue
		}

		if strings.Contains(tok, "Android") {
			*isAndroid = true
		}
		if strings.Contains(tok, "MSIE") {
			*isMSIE = true
		}

		seenBrowsers[crc32.ChecksumIEEE([]byte(tok))] = struct{}{}
		dec.More()
	}
}

func checkTokenErr(t json.Token, err error) (isBreak bool) {
	isBreak = false
	if err != nil && err != io.EOF {
		log.Println("error happened", err)
		isBreak = true
	} else if err == io.EOF { // del elif?
		log.Println("EOF")
		isBreak = true
	}
	if t == nil {
		log.Println("t is nil")
		isBreak = true
	}
	return
}
