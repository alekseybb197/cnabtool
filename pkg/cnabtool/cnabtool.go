package cnabtool

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

var webClient = http.Client{
	//Timeout: time.Millisecond * time.Duration(cfg.Timeout),
	Timeout: time.Millisecond * time.Duration(1000),
}

var media string

func webrequest(url string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	//req.Header.Set("User-Agent", cfg.Client)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept", media)

	//if cfg.Verbosity > 2 {
	fmt.Printf("Request Headers: %+v\n", req.Header)
	//}

	res, err := webClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	return res
}

func Get(repourl string) (apiresp string) {

	url := repourl
	// fix schema if needed
	schema, _ := regexp.Compile(`^http.://`) // schema mask
	if !schema.MatchString(repourl) {
		url = "http://" + repourl
	}
	//if cfg.Verbosity > 1 {
	//	fmt.Printf("Url: %+v\n", url)
	//}

	// set response accepted
	//switch cfg.Format {
	//case "json":
	media = "application/json"
	//case "text":
	//	media = "text/plain"
	//default:
	//	media = "text/html"
	//}

	//if cfg.Output != "" {
	webClient.Timeout = -1 // disable timeout
	//}

	res := webrequest(url)
	if res.Body == nil {
		log.Fatalln("Error: body is nil")
	}
	defer res.Body.Close()

	//if cfg.Verbosity > 2 {
	fmt.Printf("Response StatusCode %+v\n", res.StatusCode)
	fmt.Printf("Response Header %+v\n", res.Header)
	//}

	// catch fail code
	if res.StatusCode != 200 && res.StatusCode != 400 && res.StatusCode != 404 && res.StatusCode != 401 {
		log.Fatalf("failed to fetch data: %s", res.Status)
	}

	// choose properly context reader
	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(res.Body)
		defer reader.Close()
	default:
		reader = res.Body
	}

	bytesbody, readErr := io.ReadAll(reader)
	if readErr != nil {
		log.Fatalln(readErr)
	}
	return string(bytesbody)
}
