<p align="center">
  <a href="https://github.com/xdrm-io/aicra">
    <img src="https://github.com/xdrm-io/aicra/raw/0.4.0/readme.assets/logo.png" alt="aicra logo" width="200" height="200">
  </a>
</p>

<h3 align="center">aicra</h3>

<p align="center">
  Fast, intuitive, and powerful configuration-driven engine for faster and easier <em>REST</em> development.
</p>

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go version](https://img.shields.io/badge/go_version-1.16-blue.svg)](https://golang.org/doc/go1.16)
[![Go doc](https://pkg.go.dev/badge/github.com/xdrm-io/aicra)](https://pkg.go.dev/github.com/xdrm-io/aicra)
[![Go Report Card](https://goreportcard.com/badge/github.com/xdrm-io/aicra)](https://goreportcard.com/report/github.com/xdrm-io/aicra)
[![Build status](https://github.com/xdrm-io/aicra/actions/workflows/go.yml/badge.svg)](https://github.com/xdrm-io/aicra/actions/workflows/go.yml)
[![Coverage](https://codecov.io/gh/xdrm-io/aicra/branch/0.4.0/graph/badge.svg?token=HDIMZ0MKXW)](https://codecov.io/gh/xdrm-io/aicra)

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
  * [Parameters](#parameters)
      - [Input parameters](#input-parameters)
      - [Mandatory vs. Optional:](#mandatory-vs-optional)
      - [Renaming](#renaming)
  * [Example](#example)
- [Writing handlers](#writing-handlers)
- [Coming next](#coming-next)

<!-- tocstop -->

## Installation

To install the aicra package, you need to install Go and set your Go workspace first.
> not tested under Go 1.14

1. you can use the command below to install aicra.
```bash
$ go get -u github.com/xdrm-io/aicra
```
2. Import it in your code:
```go
import "github.com/xdrm-io/aicra"
```

## What's automated

As the configuration file is here to make your life easier, let's take a quick look at what you do not have to do ; or in other words, what does `aicra` automates.

Http requests and responses are automatically handled.

Requests are only accepted when they meet the permissions you have defined. Otherwise, the request is automatically rejected with an error.

Request data is automatically validated and extracted before it reaches your code. Missing or invalid data results in an automatic error response.

Aicra injects input data into your handlers and formats the output data back to an http response.

Any error in the configuration or your code is spotted before the server accepts incoming requests. Only when the server is valid (the configuration and your handlers), it starts listening for incoming requests. There will be no surprise at "runtime" !

Configuration errors:
- missing configuration fields
- unknown HTTP method
- invalid uri
- uri collision between 2 services
- unknown input/output data type
- collision between input/output variable names

Handler errors:
- a service from the configuration has no handler attached
- a handler from your code does not match any service from the configuration
- a handler from your code does not match the configuration service's input or output

## Getting started

Here is the minimal code to launch your aicra server assuming your configuration file is `api.json`.

```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/xdrm-io/aicra"
    "github.com/xdrm-io/aicra/api"
    "github.com/xdrm-io/aicra/validator/builtin"
)

func main() {
    builder := &aicra.Builder{}

    // add input validators
    builder.Input(validator.BoolDataType{})
    builder.Input(validator.UintDataType{})
    builder.Input(validator.StringDataType{})

    // add output types
    builder.Output("string", "")
    builder.Output("bool", true)
    builder.Output("user", UserStruct{})
    builder.Output("users", []UserStruct{})

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

    // add http middlewares (e.g. logger)
    builder.With(func(next http.Handler) http.Handler{ /* ... */ })

    // add contextual middlewares (e.g. authentication)
    builder.WithContext(func(next http.Handler) http.Handler{ /* ... */ })

    // bind handlers
    err = builder.Bind(http.MethodGet, "/user/{id}", getUserById)
    if err != nil {
        log.Fatalf("cannot bind: %s", err)
    }

    // build your services
    handler, err := builder.Build()
    if err != nil {
        log.Fatalf("cannot build: %s", err)
    }
    http.ListenAndServe("localhost:8080", handler)
}
```

If you want to use HTTPS, you can configure your own `http.Server`.
```go
func main() {
	server := &http.Server{
		Addr:      "localhost:8080",
		TLSConfig: &tls.Config{},
		// ...
		Handler: handler,
	}
	server.ListenAndServeTLS("server.crt", "server.key")
}
```

## Configuration file

First of all, the configuration uses `json`.

> Quick note if you thought: "I don't like JSON, I would have preferred yaml, or even xml !"
>
> I've had a hard time deciding and testing different formats including yaml and xml.
> But as it describes our entire api and is crucial for our server to keep working over updates; xml would have been too verbose with growth and yaml on the other side would have been too difficult to read. Json sits in the right spot for this.

Let's take a quick look at the configuration format !

> if you don't like boring explanations and prefer a working example, see [here](https://github.com/xdrm-io/articles-api/blob/main/api/definition.json)

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

### Parameters

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
            "{id}":      { "info": "...", "type": "int",     "name": "ID"    },
            "GET@title": { "info": "...", "type": "?string", "name": "Title" },
            "content":   { "info": "...", "type": "string"                   }
        },
        "out": {
            "Title":   { "info": "updated article title",   "type": "string" },
            "Content": { "info": "updated article content", "type": "string" }
        }
    }
]
```
##### Input parameters
The format of the key for input arguments defines where it comes from:
- `{var}` is an uri parameter, must be present in the `"path"`
- `GET@var` is an get parameter (see [http query](https://tools.ietf.org/html/rfc3986#section-3.4))
- `var` is body parameter

Body parameters are extracted based on the `Content-Type` http header. Available values are:
- `application/x-www-form-urlencoded`
- `multipart/form-data`
- `application/json`

##### Mandatory vs. Optional:
If you want to make a parameter optional, prefix its type with a question mark, by default all parameters are mandatory.

##### Renaming
Renaming with the field `"name"` is mandatory for:
- uri parameters, the `{var}` syntax
- get parameters, the `GET@var` syntax
- body parameters that do not start with an uppercase letter

These names are the same as input or output parameters in your code, they must begin with an uppercase letter in order to be exported and valid go.

##### Types

Every input type must match one of the input validators registered with `Builder.Input()`.
Every output type must match one of the output types registered with `Builder.Output()`

### Example
```json
[
    {
        "method": "PUT",
        "path": "/article/{id}",
        "scope": [["author"]],
        "info": "updates an article",
        "in": {
            "{id}":      { "name": "ID",      "type": "uint",    "info": "article id"          },
            "GET@title": { "name": "Title",   "type": "?string", "info": "new article title"   },
            "content":   { "name": "Content", "type": "string",  "info": "new article content" }
        },
        "out": {
            "article": { "name": "Article", "type": "article", "info": "updated article" }
        }
    }
]
```

Sample requests:

<table>
<tr>
    <td>HTTP Request</td>
    <td>ID</td>
    <td>Title</td>
    <td>Content</td>
</tr>
<tr><td>

```http
PUT /articles/26 HTTP/2
Content-Type: application/x-www-form-urlencoded

content=new content
```
</td><td>26</td><td></td><td>new content</td></tr>
<tr><td>

```http
PUT /articles/32 HTTP/2
Content-Type: multipart/form-data; boundary=XXX

--XXX
Content-Disposition: form-data; name="content"
new content
on
multiple lines
--XXX--
```
</td><td>32</td><td></td><td>new content<br>on<br>multiple lines</td></tr>
<tr><td>

```http
PUT /articles/11?title=new-title HTTP/2
Content-Type: application/json

{"content": "new content"}
```
</td><td>11</td><td>new-title</td><td>new content</td></tr>
</table>


## Writing handlers

Besides your main package where you launch your server, you will need to create handlers matching services from the configuration.

The code below implements a simple handler.
```go
// "in": {
//  "input1": { "name": "Input1", "type": "int",     "info": "..." },
//  "input2": { "name": "Input2", "type": "?string", "info": "..." }
// },
type req struct{
    Input1 int
    Input2 *string // optional are pointers
}
// "out": {
//  "output1": { "name": "Output1", "type": "string", "info": "..." },
//  "output2": { "name": "Output2", "type": "bool",   "info": "..." }
// }
type res struct{
    Output1 string
    Output2 bool
}

func myHandler(ctx context.Context, r req) (*res, error) {
    if err := fetchData(req.Input1); err != nil {
        return nil, api.ErrFailure
    }

    if req.Input2 != nil {
        if err := fetchData(req.Input2); err != nil {
            return nil, api.Error(404, "error description")
        }
    }

    return &res{Output1: "out1", Output2: true}, nil
}
```

If your handler signature does not match the configuration exactly, the server will print out the error and won't start.

The `api.Err` type automatically maps to HTTP status codes and error descriptions that will be sent to the client as json. This way, clients can manage the same format for every response:
```http
HTTP/1.1 404 OK
Content-Type: application/json

{"status":"not found"}
```

## Coming next
- [ ] support for PATCH or other random http methods. It might be interesting to generate the list of allowed methods from the configuration. A check against available http methods as a failsafe might be required.

