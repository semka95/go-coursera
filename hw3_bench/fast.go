package fast

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Name     string   `json:"name,string"`
	Email    string   `json:"email,string"`
}

func tSearch() {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	u := &User{}

	easyjson.UnmarshalFromReader(file, u)
	//fmt.Println(u)
}

func FastSearch(out io.Writer) {
	tSearch()
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(file)

	seenBrowsers := make(map[uint32]struct{})
	userNum := 0

	var foundUsers strings.Builder
	//foundUsers.Grow(1000)
	fmt.Fprint(&foundUsers, "found users:\n")

	for {
		isAndroid := false
		isMSIE := false

		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		tok, ok := t.(string)
		if ok {
			if tok == "browsers" {
				t, err := dec.Token()
				if checkTokenErr(t, err) {
					break
				}
				handleBrowser(&isAndroid, &isMSIE, dec, seenBrowsers)
				if !(isAndroid && isMSIE) {
					userNum++
					continue
				}
				handleNameEmail(dec, &foundUsers, userNum)
				userNum++
			}
		}
	}

	fmt.Fprintln(out, foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func handleNameEmail(dec *json.Decoder, foundUsers *strings.Builder, userNum int) {
	name := ""
	email := ""
	for {
		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		tok, ok := t.(string)
		if !ok {
			break
		}

		switch tok {
		case "email":
			t, err := dec.Token()
			if checkTokenErr(t, err) {
				break // this is not working here
			}

			email, ok = t.(string) // refactor
			if !ok {
				//log.Println("cant cast token to string")
				continue
			}
			email = strings.ReplaceAll(email, "@", " [at] ")
		case "name":
			t, err := dec.Token()
			if checkTokenErr(t, err) {
				break // this is not working here
			}

			name, ok = t.(string) // refactor
			if !ok {
				//log.Println("cant cast token to string")
				continue
			}
		}

	}
	fmt.Fprintf(foundUsers, "[%d] %s <%s>\n", userNum, name, email)
}

func handleBrowser(isAndroid, isMSIE *bool, dec *json.Decoder, seenBrowsers map[uint32]struct{}) {
	for {
		t, err := dec.Token()
		if checkTokenErr(t, err) {
			break
		}

		tok, ok := t.(string)
		if !ok {
			break
		}

		if strings.Contains(tok, "Android") {
			seenBrowsers[crc32.ChecksumIEEE([]byte(tok))] = struct{}{}
			*isAndroid = true
			continue
		}
		if strings.Contains(tok, "MSIE") {
			seenBrowsers[crc32.ChecksumIEEE([]byte(tok))] = struct{}{}
			*isMSIE = true
			continue
		}

	}
}

func checkTokenErr(t json.Token, err error) (isBreak bool) {
	isBreak = false
	// if err != nil && err != io.EOF {
	// 	//log.Println("error happened: ", err)
	// 	isBreak = true
	// 	return
	// }
	// if err == io.EOF {
	// 	//log.Println("EOF")
	// 	isBreak = true
	// 	return
	// }
	// if t == nil {
	// 	//log.Println("t is nil")
	// 	isBreak = true
	// }
	if err != nil || err == io.EOF || t == nil {
		isBreak = true
	}
	return
}
