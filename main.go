package main

import (
	"github.com/alknopfler/excel-to-radar-graph/pkg/web"
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.IndexHandler)
	mux.HandleFunc("/upload", web.UploadHandler)

	go web.Open("http://localhost")
	if err := http.ListenAndServe(":80", mux); err != nil {
		log.Fatal(err)
	}

}
