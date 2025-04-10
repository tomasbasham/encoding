package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tomasbasham/encoding"
	"github.com/tomasbasham/gonads"
)

type FormRequest struct {
	Name    string                    `form:"name"`
	Age     int                       `form:"age"`
	Aliases gonads.Optional[[]string] `form:"aliases"`
}

func main() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("name=john&age=20&aliases=jonny&aliases=johnny"))

	// Handle the request
	handleRequest(w, r)

	// Print the response
	fmt.Println(w.Body.String())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var req FormRequest

	dec := encoding.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Name: %s\n", req.Name)
	fmt.Fprintf(w, "Age: %d\n", req.Age)

	if req.Aliases.IsNone() {
		fmt.Fprint(w, "Aliases: empty\n")
		return
	}
	fmt.Fprintf(w, "Aliases: %v", req.Aliases.Unwrap())
}
