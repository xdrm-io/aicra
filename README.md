<p align="center">
  <a href="https://git.xdrm.io/go/aicra">
    <img src="https://git.xdrm.io/go/aicra/raw/branch/feature/improve-readme/readme.assets/logo.png" alt="aicra logo" width="200" height="200">
  </a>
</p>

<h3 align="center">aicra</h3>

<p align="center">
  Fast, intuitive, and powerful configuration-driven engine for faster and easier <em>REST</em> development.
</p>

[![Go version](https://img.shields.io/badge/go_version-1.16-blue.svg)](https://golang.org/doc/go1.16) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/git.xdrm.io/go/aicra)](https://goreportcard.com/report/git.xdrm.io/go/aicra) [![Go doc](https://godoc.org/git.xdrm.io/go/aicra?status.svg)](https://godoc.org/git.xdrm.io/go/aicra) [![Build Status](https://drone.xdrm.io/api/badges/go/aicra/status.svg)](https://drone.xdrm.io/go/aicra)

## Presentation

`aicra` is a lightweight and idiomatic configuration-driven engine for building REST services. It's especially good at helping you write large APIs that remain maintainable as your project grows.

The focus of the project is to allow you to build a fully-featured REST API in an elegant, comfortable and inexpensive way. This is achieved by using a single configuration file to drive the server. This one file describes your entire API: methods, uris, input data, expected output, permissions, etc.

Repetitive tasks are automatically processed by `aicra` based on your configuration, you're left with implementing your handlers (_usually business logic_).

## Table of contents

<!-- toc -->

- [Installation](#installation)
- [What's automated](#whats-automated)
- [Getting started](#getting-started)
- [Configuration file](#configuration-file)
  * [Services](#services)
  * [Input and output parameters](#input-and-output-parameters)
  * [Example](#example)
- [Writing your code](#writing-your-code)
- [Changelog](#changelog)

<!-- tocstop -->

## Installation

To install the aicra package, you need to install Go and set your Go workspace first.
> not tested under Go 1.14

1. you can use the below Go command to install aicra.
```bash
$ go get -u git.xdrm.io/go/aicra
```
2. Import it in your code:
```go
import "git.xdrm.io/go/aicra"
```

## What's automated

As the configuration file is here to make your life easier, let's take a quick look at what you do not have to do ; or in other words, what does `aicra` automates.

Http requests are only accepted when they have the permissions you have defined. If unauthorized, the request is rejected with an error response.

Request data is automatically extracted and validated before it reaches your code. If a request has missing or invalid data an automatic error response is sent.

When launching the server, it ensures everything is ok and won't start until fixed. You will get errors for:
- handler signature does not match the configuration
- a configuration service has no handler
- a handler does not match any service

The same applies if your configuration is invalid:
- unknown HTTP method
- invalid uri
- uri collision between 2 services
- missing fields
- unknown data type
- input name collision

## Getting started

Here is the minimal code to launch your aicra server assuming your configuration file is `api.json`.

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

    // register data validators
    builder.AddType(builtin.BoolDataType{})
    builder.AddType(builtin.UintDataType{})
    builder.AddType(builtin.StringDataType{})

    // load your configuration
    config, err := os.Open("api.json")
    if err != nil {
        log.Fatalf("cannot open config: %s", err)
    }
    err = builder.Setup(config)
    config.Close() // free config file
    if err != nil {
        log.Fatalf("invalid config: %s", err)
    }

    // bind handlers
    err = builder.Bind(http.MethodGet, "/user/{id}", getUserById)
    if err != nil {
        log.Fatalf("cannog bind GET /user/{id}: %s", err)
    }
    // ...

    // build your services
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
        // ...
		Handler:   AICRAHandler,
	}
    server.ListenAndServe()
}
```

## Configuration file

First of all, the configuration uses `json`.

> Quick note if you thought: "I hate JSON, I would have preferred yaml, or even xml !"
>
> I've had a hard time deciding and testing different formats including yaml and xml.
> But as it describes our entire api and is crucial for our server to keep working over updates; xml would have been too verbose with growth and yaml on the other side would have been too difficult to read. Json sits in the right spot for this.

Let's take a quick look at the configuration format !

> if you don't like boring explanations and prefer a working example, see [here](https://git.xdrm.io/go/articles-api/src/master/api.json)

### Services

To begin with, the configuration file defines a list of services. Each one is defined by:
- `method` an HTTP method
- `path` an uri pattern (can contain variables)
- `info` a short description of what it does
- `scope` a list of the required permissions
- `in` a list of input arguments
- `out` a list of output arguments
```json
[
    {
        "method": "GET",
        "path": "/article",
        "scope": [["author", "reader"], ["admin"]],
        "info": "returns all available articles",
        "in": {},
        "out": {}
    }
]
```

The `scope` is a 2-dimensional list of permissions. The first list means **or**, the second means **and**, it allows for complex permission combinations. The example above can be translated to: this method requires users to have permissions (author **and** reader) **or** (admin)

### Input and output parameters

Input and output parameters share the same format, featuring:
- `info` a short description of what it is
- `type` its data type (_c.f. validation_)
- `?` whether it is mandatory or optional
- `name` a custom name for easy access in code
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
            "title":   { "info": "updated article title",   "type": "string" },
            "content": { "info": "updated article content", "type": "string" }
        }
    }
]
```

If a parameter is optional you just have to prefix its type with a question mark, by default all parameters are mandatory.

The format of the key of input arguments defines where it comes from:
1. `{param}` is an URI parameter that is extracted from the `"path"`
2. `GET@param` is an URL parameter that is extracted from the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
3. `param` is a body parameter that can be extracted from 3 formats independently:
    - _url encoded_: data send in the body following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
    - _multipart_: data send in the body with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format can be quite heavy but allows to transmit data as well as files.
    - _JSON_: data sent in the body as a json object ; The _Content-Type_ header must be `application/json` for it to work.

### Example
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

1. `{id}` is extracted from the end of the URI and is a number compliant with the `int` type checker. It is renamed `ID`, this new name will be sent to the handler.
2. `GET@title` is extracted from the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `string` or not given at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `title`.
3. `content` can be extracted from json, multipart or url-encoded data; it makes no difference and only give clients a choice over the technology to use. If not renamed, the variable will be given to the handler with its original name `content`.



## Writing your code

Besides your main package where you launch your server, you will need to create handlers matching services from the configuration.

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

func myHandler(ctx *api.Context, r req) (*res, api.Err) {
    err := doSomething()
    if err != nil {
        return nil, api.ErrFailure
    }
    return &res{"out1", true}, api.ErrSuccess
}
```

If your handler signature does not match the configuration exactly, the server will print out the error and won't start.

The `api.Err` type automatically maps to HTTP status codes and error descriptions that will be sent to the client as json; clients have to manage the same format for every response.
```json
{
    "error": {
        "code": 0,
        "reason": "all right"
    }
}
```

## Changelog

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

