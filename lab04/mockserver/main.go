package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Got request %s\n", r.Method)
		switch r.Method {
		case http.MethodGet:
			fmt.Printf("Sending %d\n", http.StatusNotFound)
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPost:
			defer r.Body.Close()
			buf, _ := io.ReadAll(r.Body)
			fmt.Printf("Sending %d\n", http.StatusOK)
			w.WriteHeader(http.StatusOK)
			fmt.Printf("Sending body:\n%s\n", string(buf))
			w.Write(buf)
		}
	})

	http.ListenAndServe(":8000", mux)
}
