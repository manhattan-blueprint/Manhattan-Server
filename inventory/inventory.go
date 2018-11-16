package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Inventory!\n"))
}

func main() {
	fmt.Println("hello inventory")
	r := mux.NewRouter()
	r.HandleFunc("/", TestHandler)
	log.Fatal(http.ListenAndServe(":8000", r))
}
