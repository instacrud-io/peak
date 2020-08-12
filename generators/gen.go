// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates contributors.go. It can be invoked by running
// go generate
package main

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	const url = "https://github.com/golang/go/raw/master/CONTRIBUTORS"

	rsp, err := http.Get(url)
	die(err)
	defer rsp.Body.Close()

	sc := bufio.NewScanner(rsp.Body)
	carls := []string{}

	for sc.Scan() {
		if strings.Contains(sc.Text(), "Carl") {
			carls = append(carls, sc.Text())
		}
	}

	die(sc.Err())

	f, err := os.Create("contributors.go")
	die(err)
	defer f.Close()

	packageTemplate := template.Must(template.ParseFiles("templates/dummy.tmpl"))
	//template.ParseFiles("templates/dummy.tmpl")

	packageTemplate.Execute(f, struct {
		Timestamp time.Time
		URL       string
		Model     string
		Carls     []string
	}{
		Timestamp: time.Now(),
		URL:       url,
		Carls:     carls,
		Model:     "mestri",
	})
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
