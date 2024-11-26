package main

import (
	"bytes"
	"fmt"

	// "regexp"
	regexp "github.com/wasilibs/go-re2"
)

func main() {

	var parterns = make(map[string]string)
	parterns["email"] = `(?P<name>[-\w\d\.]+?)(?:\s+at\s+|\s*@\s*|\s*(?:[\[\]@]){3}\s*)(?P<host>[-\w\d\.]*?)\s*(?:dot|\.|(?:[\[\]dot\.]){3,5})\s*(?P<domain>\w+)`
	parterns["bitcoin"] = `\b([13][a-km-zA-HJ-NP-Z1-9]{25,34}|bc1[ac-hj-np-zAC-HJ-NP-Z02-9]{11,71})`
	parterns["ssn"] = `\d{3}-\d{2}-\d{4}`
	parterns["uri"] = `[\w]+://[^/\s?#]+[^\s?#]+(?:\?[^\s#]*)?(?:#[^\s]*)?`
	parterns["tel"] = `\+\d{1,4}?[-.\s]?\(?\d{1,3}?\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}`

	var data = bytes.Repeat([]byte("123@mail.co nümbr=+71112223334 SSN:123-45-6789 http://1.1.1.1 3FZbgi29cpjq2GjdwV8eyHuJJnkLtktZc5 Й"), 100)

	var partern = ""
	for key, value := range parterns {
		if len(partern) > 0 {
			partern += "|"
		}
		partern += fmt.Sprintf("(?P<%s>%s)", key, value)
	}
	fmt.Println("partern: ", partern)
	// fmt.Printf("data: %s\n", data)

	re, err := regexp.Compile(partern)
	if err != nil {
		panic(err)
	}

	loop := 10000 // 1000000
	match_count := 0
	for i := 0; i < loop; i++ {

		// Use the regular expression to find matches in a string.
		matches := re.FindAllSubmatch(data, -1) //.FindSubmatch(data)
		if matches == nil {
			// fmt.Println("No match found")
		} else {
			match_count += len(matches)
			if i == 0 {
				// fmt.Printf("Match found: %d, %#v\n", len(matches), matches)
				for j, match := range matches {
					fmt.Printf("\t%d Match found: %s\n", j, match)
				}
			}
			if i%1000 == 0 {
				fmt.Printf("Match with loop %d count: %d\n", i, match_count)
			}
		}
	}

	fmt.Printf("Match with loop %d count: %d", loop, match_count)
}
