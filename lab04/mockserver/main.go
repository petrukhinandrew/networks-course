package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPost:
			defer r.Body.Close()
			buf, err := io.ReadAll(r.Body)
			fmt.Printf("%d, %s, %v\n", len(buf), string(buf), err)
			w.Header()["Content-Length"] = []string{fmt.Sprintf("%d", len(buf))}
			w.Header()["Content-Type"] = []string{"text/plain", "enctype=utf-8"}
			w.WriteHeader(http.StatusOK)
			w.Write(buf)

		}
	})

	http.ListenAndServe(":8000", mux)
}
