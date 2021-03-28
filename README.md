# | aicra |

[![Go version](https://img.shields.io/badge/go_version-1.10.3-blue.svg)](https://golang.org/doc/go1.10)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/git.xdrm.io/go/aicra)](https://goreportcard.com/report/git.xdrm.io/go/aicra)
[![Go doc](https://godoc.org/git.xdrm.io/go/aicra?status.svg)](https://godoc.org/git.xdrm.io/go/aicra)
[![Build Status](https://drone.xdrm.io/api/badges/go/aicra/status.svg)](https://drone.xdrm.io/go/aicra)

----

`aicra` is a lightweight and idiomatic API engine for building Go services. It's especially good at helping you write large REST API services that remain maintainable as your project grows.

The focus of the project is to allow you to build a fully featured REST API in an elegant, comfortable and inexpensive way. This is achieved by using a configuration file to drive the server. The configuration format describes the whole API: routes, input arguments, expected output, permissions, etc.

TL;DR: `aicra` is a fast configuration-driven REST API engine.

Repetitive tasks is automatically processed by `aicra` from your configuration file, you just have to implement your handlers.

The engine automates :
- catching input data (_url, query, form-data, json, url-encoded_)
- handling missing input data (_required arguments_)
- handling input data validation
- checking for mandatory output parameters
- checking for missing method implementations
- checking for handler signature (input and output arguments)

> An example project is available [here](https://git.xdrm.io/go/articles-api)

### Table of contents

<!-- toc -->

  * [Installation](#installation)
- [Usage](#usage)
      - [Create a server](#create-a-server)
      - [Create a handler](#create-a-handler)
- [Configuration](#configuration)
      - [Global format](#global-format)
        * [Input section](#input-section)
          + [Format](#format)
      - [Example](#example)
- [Changelog](#changelog)

<!-- tocstop -->

## Installation

You need a recent machine with `go` [installed](https://golang.org/doc/install). The package has not been tested under **go1.14**.


```bash
go get -u git.xdrm.io/go/aicra
```


# Usage


#### Create a server

The code below sets up and creates an HTTP server from the `api.json` configuration.

```go
package main

import (
    "log"
    "net/http"
    "os"

    "git.xdrm.io/go/aicra"
    "git.xdrm.io/go/aicra/api"
    "git.xdrm.io/go/aicra/datatype/builtin"
)

func main() {
    builder := &aicra.Builder{}

    // register available validators
    builder.AddType(builtin.BoolDataType{})
    builder.AddType(builtin.UintDataType{})
    builder.AddType(builtin.StringDataType{})

    // load your configuration
    config, err := os.Open("./api.json")
    if err != nil {
        log.Fatalf("cannot open config: %s", err)
    }
    err = builder.Setup(config)
    config.Close() // free config file
    if err != nil {
        log.Fatalf("invalid config: %s", err)
    }

    // bind your handlers
    builder.Bind(http.MethodGet, "/user/{id}", getUserById)
    builder.Bind(http.MethodGet, "/user/{id}/username", getUsernameByID)

    // build the handler and start listening
    handler, err := builder.Build()
    if err != nil {
        log.Fatalf("cannot build handler: %s", err)
    }
    http.ListenAndServe("localhost:8080", handler)
}
```

If you want to use HTTPS, you can configure your own `http.Server`.

```go
func main() {
    server := &http.Server{
        Addr:      "localhost:8080",
		TLSConfig: tls.Config{},
        // aicra handler
		Handler:   handler,
	}

    server.ListenAndServe()
}
```


#### Create a handler

The code below implements a simple handler.
```go
// "in": {
//  "Input1": { "info": "...", "type": "int"     },
//  "Input2": { "info": "...", "type": "?string" }
// },
type req struct{
    Input1 int
    Input2 *string // optional are pointers
}
// "out": {
//  "Output1": { "info": "...", "type": "string" },
//  "Output2": { "info": "...", "type": "bool"   }
// }
type res struct{
    Output1 string
    Output2 bool
}

func myHandler(r req) (*res, api.Err) {
    err := doSomething()
    if err != nil {
        return nil, api.ErrFailure
    }
    return &res{"out1", true}, api.ErrSuccess
}
```

If your handler signature does not match the configuration exactly, the server will print out the error and will not start.

The `api.Err` type automatically maps to HTTP status codes and error descriptions that will be sent to the client as json; client will then always have to manage the same format.
```json
{
    "error": {
        "code": 0,
        "reason": "all right"
    }
}
```


# Configuration

The whole api behavior is described inside a json file (_e.g. usually api.json_). For a better understanding of the format, take a look at this working [configuration](https://git.xdrm.io/go/articles-api/src/master/api.json).

The configuration file defines :
- routes and their methods
- every input argument for each method
- every output for each method
- scope permissions (list of permissions required by clients)
- input policy :
    - type of argument (_c.f. data types_)
    - required/optional
    - variable renaming

#### Global format

The root of the json file must feature an array containing your requests definitions. For each, you will have to create fields described in the table above.

- `info`: Short description of the method
- `in`: List of arguments that the clients will have to provide. [Read more](#input-arguments).
- `out`: List of output data that your controllers will output. It has the same syntax as the `in` field but optional parameters are not allowed.
- `scope`: A 2-dimensional array of permissions. The first level means **or**, the second means **and**. It allows to combine permissions in complex ways.
    - Example: `[["A", "B"], ["C", "D"]]` translates to : this method requires users to have permissions (A **and** B) **or** (C **and** D)


##### Input section

Input arguments defines what data from the HTTP request the method requires. `aicra` is able to extract 3 types of data :

- **URI** - data from inside the request path. For instance, if your controller is bound to the `/user/{id}` URI, you can set the input argument `{id}` matching this uri part.
- **Query** - data at the end of the URL following the standard [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **Form** - data send from the body of the request ; it can be extracted in 3 ways:
    - _URL encoded_: data send in the body following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
    - _Multipart_: data send in the body with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format can be quite heavy but allows to transmit data as well as files.
    - _JSON_: data send in the body as a json object ; each key being a variable name, each value its content. Note that the 'Content-Type' header must be set to `application/json` for the API to use it.

> For Form data, the 3 methods can be used at once for different arguments; for instance if you need to send a file to an aicra server as well as other parameters, you can use JSON for parameters and Multipart for the file.

###### Format

The `in` field describes as list of arguments where the key is the argument name, and the value defines how to manage the variable.


Variable names from **URI** or **Query** must be named accordingly :
- an **URI** variable `{var}` from your request route must be named `{var}` in the `in` section
- a variable `var` in the **Query** has to be named `GET@var` in the `in` section


#### Example
```json
[
    {
        "method": "PUT",
        "path": "/article/{id}",
        "scope": [["author"]],
        "info": "updates an article",
        "in": {
            "{id}":      { "info": "...", "type": "int",     "name": "id"    },
            "GET@title": { "info": "...", "type": "?string", "name": "title" },
            "content":   { "info": "...", "type": "string"                   }
        },
        "out": {
            "id":      { "info": "updated article id",      "type": "uint"   },
            "title":   { "info": "updated article title",   "type": "string" },
            "content": { "info": "updated article content", "type": "string" }
        }
    }
]
```

1. `{id}` is extracted from the end of the URI and is a number compliant with the `int` type checker. It is renamed `id`, this new name will be sent to the handler.
2. `GET@title` is extracted from the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `string` or not given at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `title`.
3. `content` can be extracted from json, multipart or url-encoded data; it makes no difference and only give clients a choice over the technology to use. If not renamed, the variable will be given to the handler with its original name `content`.


# Changelog

- [x] human-readable json configuration
- [x] nested routes (*i.e. `/user/{id}` and `/user/post/{id}`*)
- [x] nested URL arguments (*i.e. `/user/{id}` and `/user/{uid}/post/​{id}`*)
- [x] useful http methods: GET, POST, PUT, DELETE
    - [ ] add support for PATCH method
    - [ ] add support for OPTIONS method
        - [ ] it might be interesting to generate the list of allowed methods from the configuration
        - [ ] add CORS support
- [x] manage request data extraction:
    - [x] URL slash-separated strings
    - [x] HTTP Query named parameters
        - [x] manage array format
    - [x] body parameters
        - [x] multipart/form-data (variables and file uploads)
        - [x] application/x-www-form-urlencoded
        - [x] application/json
- [x] required vs. optional parameters with a default value
- [x] parameter renaming
- [x] generic type check (*i.e. you can add custom types alongside built-in ones*)
- [x] built-in types
    - [x] `any` - matches any value
    - [x] `int` - see go types
    - [x] `uint` - see go types
    - [x] `float` - see go types
    - [x] `string` - any text
    - [x] `string(len)` - any string with a length of exactly `len` characters
    - [x] `string(min, max)` - any string with a length between `min` and `max`
    - [ ] `[]a` - array containing **only** elements matching `a` type
    - [ ] `a[b]` - map containing **only** keys of type `a` and values of type `b` (*a or b can be ommited*)
- [x] generic handler implementation
- [x] response interface
- [x] generic errors that automatically formats into response
    - [x] builtin errors
    - [x] possibility to add custom errors
- [x] check for missing handlers when building the handler
- [x] check handlers not matching a route in the configuration at server boot
- [x] specific configuration format errors qt server boot
- [x] statically typed handlers - avoids having to check every input and its type (_which is used by context.Context for instance_)
    - [x] using reflection to use structs as input and output arguments to match the configuration
        - [x] check for input and output arguments structs at server boot
- [x] check for unavailable types in configuration at server boot
- [x] recover panics from handlers
- [ ] improve tests and coverage

