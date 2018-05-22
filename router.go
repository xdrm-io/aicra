package gfw

import (
	"fmt"
	"git.xdrm.io/gfw/internal/config"
	"log"
	"net/http"
	"strings"
)

func (s Server) route(res http.ResponseWriter, req *http.Request) {

	/* (1) Build request
	---------------------------------------------------------*/
	/* (1) Try to build request */
	request, err := buildRequest(req)
	if err != nil {
		log.Fatal(req)
	}

	/* (2) Find a controller
	---------------------------------------------------------*/
	/* (1) Init browsing cursors */
	ctl := s.config
	uriIndex := 0

	/* (2) Browse while there is uri parts */
	for uriIndex < len(request.Uri) {
		uri := request.Uri[uriIndex]

		child, hasKey := ctl.Children[uri]

		// stop if no matchind child
		if !hasKey {
			break
		}

		request.ControllerUri = append(request.ControllerUri, uri)
		ctl = child
		uriIndex++

	}

	/* (3) Extract URI params */
	uriParams := request.Uri[uriIndex:]

	/* (4) Store them as Data */
	for i, data := range uriParams {
		request.UrlData = append(request.UrlData, data)
		request.Data[fmt.Sprintf("URL#%d", i)] = data
	}

	/* (3) Check method
	---------------------------------------------------------*/
	/* (1) Unavailable method */
	if req.Method == "GET" && ctl.GET == nil ||
		req.Method == "POST" && ctl.POST == nil ||
		req.Method == "PUT" && ctl.PUT == nil ||
		req.Method == "DELETE" && ctl.DELETE == nil {

		Json, _ := ErrUnknownMethod.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", ErrUnknownMethod.Reason)
		return

	}

	/* (2) Extract method cursor */
	var method *config.Method

	if req.Method == "GET" {
		method = ctl.GET
	} else if req.Method == "POST" {
		method = ctl.POST
	} else if req.Method == "PUT" {
		method = ctl.PUT
	} else if req.Method == "DELETE" {
		method = ctl.DELETE
	}

	/* (3) Unmanaged HTTP method */
	if method == nil { // unknown method
		Json, _ := ErrUnknownMethod.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", ErrUnknownMethod.Reason)
		return
	}

	/* (4) Check arguments
	---------------------------------------------------------*/
	for name, data := range request.Data {
		fmt.Printf("- %s: %v\n", name, data)
	}
	fmt.Printf("\n")

	fmt.Printf("OK\nplugin: '%si.so'\n", strings.Join(request.ControllerUri, "/"))
	return
}
