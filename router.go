package gfw

import (
	"fmt"
	"log"
	"net/http"
)

func route(res http.ResponseWriter, req *http.Request) {

	/* (1) Build request
	---------------------------------------------------------*/
	/* (1) Try to build request */
	request, err := buildRequest(req)
	if err != nil {
		log.Fatal(req)
	}

	fmt.Printf("Uri:    %v\n", request.Uri)
	fmt.Printf("GET:    %v\n", request.GetData)
	fmt.Printf("POST:   %v\n", request.FormData)

	// 1. Query parameters
	// fmt.Printf("query: %v\n", req.URL.Query())

	// 2. URI path
	// fmt.Printf("uri: %v\n", req.URL.Path)

	// 3. Form values
	// fmt.Printf("form: %v\n", req.FormValue("asa"))

}
