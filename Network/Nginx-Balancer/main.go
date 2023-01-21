package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	argsRaw := os.Args
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from", argsRaw[2])
	})

	http.ListenAndServe(argsRaw[1], nil)
}
