package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var config Configuration

func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Inventory!\n")
}

func main() {
	fmt.Println("hello inventory")

	// Get the configuration
	config, err := GetConfiguration("conf.json")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", TestHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(config.Port)),
		r))
}
