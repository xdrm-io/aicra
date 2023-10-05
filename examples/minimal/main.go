package main

import (
	"log"
	"net/http"
	"time"

	_ "embed"

	"github.com/xdrm-io/aicra/examples/minimal/generated"
	"github.com/xdrm-io/aicra/runtime"
)

func main() {
	var (
		db        = &DB{}
		endpoints = NewEndpoints(db)
	)

	builder, err := generated.New(endpoints)
	if err != nil {
		log.Fatalf("cannot setup builder: %s", err)
	}

	// add http middlewares (logger)
	builder.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s '%s' in %s", r.Method, r.RequestURI, time.Now().Sub(start).String())
		})
	})

	// add contextual middlewares (authentication)
	builder.WithContext(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := runtime.GetAuth(r)
			if auth == nil {
				panic("auth is unavailable")
			}
			auth.Active = append(auth.Active, r.Header.Get("Authorization"))
			next.ServeHTTP(w, r)
		})
	})

	// build your api
	handler, err := builder.Build(generated.Validators)
	if err != nil {
		log.Fatalf("cannot build handler: %s", err)
	}
	log.Printf("server up at ':8080'")
	http.ListenAndServe("localhost:8080", handler)
}
