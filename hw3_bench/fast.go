package fast

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Name     string   `json:"name,string"`
	Email    string   `json:"email,string"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	u := &User{}

	var foundUsers strings.Builder
	foundUsers.Grow(5000)
	fmt.Fprint(&foundUsers, "found users:\n")

	seenBrowsers := make(map[string]struct{})
	userNum := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		isAndroid := false
		isMSIE := false
		u.UnmarshalJSON(scanner.Bytes())

		for _, browser := range u.Browsers {
			if ok := strings.Contains(browser, "Android"); ok {
				seenBrowsers[browser] = struct{}{}
				isAndroid = true
				continue
			}
			if ok := strings.Contains(browser, "MSIE"); ok {
				seenBrowsers[browser] = struct{}{}
				isMSIE = true
				continue
			}
		}

		if !(isAndroid && isMSIE) {
			userNum++
			continue
		}
		fmt.Fprintf(&foundUsers, "[%d] %s <%s>\n", userNum, u.Name, strings.ReplaceAll(u.Email, "@", " [at] "))
		userNum++
	}

	fmt.Fprintln(out, foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func OldFastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(file)

	seenBrowsers := make(map[string]struct{})
	userNum := 0

	var foundUsers strings.Builder
	foundUsers.Grow(1000)
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

func handleBrowser(isAndroid, isMSIE *bool, dec *json.Decoder, seenBrowsers map[string]struct{}) {
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
			seenBrowsers[tok] = struct{}{}
			*isAndroid = true
			continue
		}
		if strings.Contains(tok, "MSIE") {
			seenBrowsers[tok] = struct{}{}
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
