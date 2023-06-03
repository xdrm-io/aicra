package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/validator"
)

// config defines the API with one endpoint in this case
//   - description:   updates an existing user by defining its username, firstname,
//     and lastname   in any combination
//   - http method:   PUT
//   - http uri:      /user/{id} where {id} is a uint
//   - http request:  3 body parameters: username, firstname, lastname, all are
//     optional and are strings with a size between 3 and 20
//     characters (included)
//   - permissions:   `admin` or the requested user that has the permission `user[{id}]`
//     e.g. the user 123 has the permission `user[123]` when logged in
//     e.g. it can access this endpoint with /user/123 but not /user/456
//   - http response: the 3 updated fields as json username, firstname, lastname
const config string = `[
	{
		"method": "PUT",
		"path": "/user/{id}",
		"scope": [ ["admin"], ["user[ID]"] ],
		"info": "updates user information",
		"in": {
			"{id}":        { "info": "id of the user to update",    "type": "uint",          "name": "ID"        },
			"GET@dry_run": { "info": "whether to dry-run the call", "type": "?bool",         "name": "DryRun"    },
			"username":    { "info": "optional new username",       "type": "?string(3,20)", "name": "Username"  },
			"firstname":   { "info": "optional new firstname",      "type": "?string(3,20)", "name": "Firstname" },
			"lastname":    { "info": "optional new lastname",       "type": "?string(3,20)", "name": "Lastname"  }
		},
		"out": {
			"username":  { "info": "new username",  "type": "string(3,20)", "name": "Username"  },
			"firstname": { "info": "new firstname", "type": "string(3,20)", "name": "Firstname" },
			"lastname":  { "info": "new lastname",  "type": "string(3,20)", "name": "Lastname"  }
		}
	}
]`

func main() {
	builder := &aicra.Builder{}

	// add custom type validators
	builder.Input(validator.BoolType{})
	builder.Input(validator.UintType{})
	builder.Input(validator.StringType{})

	// load your configuration
	err := builder.Setup(strings.NewReader(config))
	if err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	endpoints := &Endpoints{db: &DB{}}

	// create user 1
	_, err = endpoints.db.CreateUser("username", "firstname", "lastname")
	if err != nil {
		log.Fatalf("cannot create user 1: %s", err)
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

	// bind handlers
	err = aicra.Bind(builder, http.MethodPut, "/user/{id}", endpoints.updateUser)
	if err != nil {
		log.Fatalf("cannot bind GET /user/{id}: %s", err)
	}

	// build your services
	handler, err := builder.Build()
	if err != nil {
		log.Fatalf("cannot build handler: %s", err)
	}
	log.Printf("server up at ':8080'")
	http.ListenAndServe("localhost:8080", handler)
}
