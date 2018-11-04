package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Authenticate!\n"))
}

func main() {
	fmt.Println("hello authenticate")
	r := mux.NewRouter()
	r.HandleFunc("/", TestHandler)
	log.Fatal(http.ListenAndServe(":8000", r))
}
