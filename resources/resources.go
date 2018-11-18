package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var config = GetConfiguration("conf.json")

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Resources!\n"))
}

func main() {
	fmt.Println("hello resources")
	r := mux.NewRouter()
	r.HandleFunc("/", TestHandler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), r))
}
