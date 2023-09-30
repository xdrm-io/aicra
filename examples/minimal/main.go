package main

import (
	"log"
	"net/http"
	"time"

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/examples/minimal/generated"
)

func main() {
	var (
		db        = &DB{}
		endpoints = NewEndpoints(db)
	)

	builder := &aicra.Builder{}
	err := generated.Bind(builder, endpoints)
	if err != nil {
		log.Fatalf("cannot wire server: %s", err)
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
			ctx := api.Extract(r.Context())
			if ctx == nil {
				panic("ctx is unavailable")
			}
			ctx.Auth.Active = append(ctx.Auth.Active, "user[1]")
			next.ServeHTTP(w, r)
		})
	})

	// build your services
	handler, err := builder.Build()
	if err != nil {
		log.Fatalf("cannot build handler: %s", err)
	}
	log.Printf("server up at ':8080'")
	http.ListenAndServe("localhost:8080", handler)
}
