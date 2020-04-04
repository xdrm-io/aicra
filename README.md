# | aicra |

[![Go version](https://img.shields.io/badge/go_version-1.10.3-blue.svg)](https://golang.org/doc/go1.10)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/git.xdrm.io/go/aicra)](https://goreportcard.com/report/git.xdrm.io/go/aicra)
[![Go doc](https://godoc.org/git.xdrm.io/go/aicra?status.svg)](https://godoc.org/git.xdrm.io/go/aicra)
[![Build Status](https://drone.xdrm.io/api/badges/go/aicra/status.svg)](https://drone.xdrm.io/go/aicra)


Aicra is a *configuration-driven* REST API engine written in Go.

Most of the management is done for you using a configuration file describing your API. you're left with implementing :
- handlers
- optionnally middle-wares (_e.g. authentication, csrf_)
- and optionnally your custom type checkers to check input parameters

> A example project is available [here](https://git.xdrm.io/go/articles-api)

## Table of contents

<!-- toc -->

- [I/ Installation](#i-installation)
- [II/ Usage](#ii-usage)
  * [1) Build a server](#1-build-a-server)
  * [2) API Configuration](#2-api-configuration)
      - [Definition](#definition)
    + [Input Arguments](#input-arguments)
      - [1. Input types](#1-input-types)
      - [2. Global Format](#2-global-format)
- [III/ Change Log](#iii-change-log)

<!-- tocstop -->

## I/ Installation

You need a recent machine with `go` [installed](https://golang.org/doc/install). This package has not been tested under the version **1.14**.


```bash
go get -u git.xdrm.io/go/aicra/cmd/aicra
```

The library should now be available as `git.xdrm.io/go/aicra` in your imports.


## II/ Usage


### 1) Build a server

Here is some sample code that builds and sets up an aicra server using your api configuration file.

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

	// add datatypes your api uses
	builder.AddType(builtin.BoolDataType{})
	builder.AddType(builtin.UintDataType{})
	builder.AddType(builtin.StringDataType{})

	config, err := os.Open("./api.json")
	if err != nil {
		log.Fatalf("cannot open config: %s", err)
	}

	// pass your configuration
	err = builder.Setup(config)
	config.Close()
	if err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	// bind your handlers
	builder.Bind(http.MethodGet, "/user/{id}", getUserById)
	builder.Bind(http.MethodGet, "/user/{id}/username", getUsernameByID)

	// build the server and start listening
	server, err := builder.Build()
	if err != nil {
		log.Fatalf("cannot build server: %s", err)
	}
	http.ListenAndServe("localhost:8080", server)
}
```


Here is an example handler
```go
type req struct{
	Param1 int
	Param3 *string // optional are pointers
}
type res struct{
	Output1 string
	Output2 bool
}

func myHandler(r req) (*res, api.Error) {
	err := doSomething()
	if err != nil {
		return nil, api.ErrorFailure
	}
	return &res{}, api.ErrorSuccess
}
```


### 2) API Configuration

The whole api behavior is described inside a json file (_e.g. usually api.json_). For a better understanding of the format, take a look at this working [template](https://git.xdrm.io/go/articles-api/src/master/api.json). This file defines :

- routes and their methods
- every input for each method (called *argument*)
- every output for each method
- scope permissions (list of permissions needed by clients)
- input policy :
	- type of argument (_c.f. data types_)
	- required/optional
	- variable renaming

#### Format

The root of the json file must be an array containing your requests definitions. For each, you will have to create fields described in the table above.

| field path | description                                                  | example                                                      |
| ---------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| `info`     | A short human-readable description of what the method does   | `create a new user`                                          |
| `scope`    | A 2-dimensional array of permissions. The first dimension can be translated to a **or** operator, the second dimension as a **and**. It allows you to combine permissions in complex ways. | `[["A", "B"], ["C", "D"]]` can be translated to : this method needs users to have permissions (A **and** B) **or** (C **and** D) |
| `in`       | The list of arguments that the clients will have to provide. [Read more](#input-arguments). | |
| `out`      | The list of output data that will be returned by your controllers. It has the same syntax as the `in` field but optional parameters are not allowed  |


### Input Arguments

Input arguments defines what data from the HTTP request the method needs. Aicra is able to extract 3 types of data :

- **URI** - data from inside the request path. For instance, if your controller is bound to the `/user/{id}` URI, you can set the input argument `{id}` matching this uri part.
- **Query** - data formatted at the end of the URL following the standard [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **URL encoded** - data send inside the body of the request but following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **Multipart** - data send inside the body of the request with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format is not very lightweight but allows you to receive data as well as files.
- **JSON** - data send inside the body as a json object ; each key being a variable name, each value its content. Note that the HTTP header '**Content-Type**' must be set to `application/json` for the API to use it.



#### Format

The `in` field in each method contains as list of arguments where the key is the argument name, and the value defines how to manage the variable.

> Variable names from **URI** or **Query** must be named accordingly :
>
> - the **URI** variable `{id}` from your request route must be named `{id}`.
> - the variable `somevar` in the **Query** has to be names `GET@somevar`.

**Example**

In this example we want 3 arguments :

```json
[
	{
		"method": "PUT",
		"path": "/article/{id}",
		"scope": [["author"]],
		"info": "updates an article",
		"in": {
			"{id}":      { "info": "article id",          "type": "int",     "name": "article_id" },
			"GET@title": { "info": "new article title",   "type": "?string", "name": "title"      },
			"content":   { "info": "new article content", "type": "string"                        }
		},
		"out": {
			"id":      { "info": "updated article id",      "type": "uint"   },
			"title":   { "info": "updated article title",   "type": "string" },
			"content": { "info": "updated article content", "type": "string" }
		}
	}
]
```

- the 1^st^ one is send at the end of the URI and is a number compliant with the `int` type checker. It is renamed `article_id`, this new name will be sent to the handler.
- the 2^nd^ one is send in the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `string` or not given at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `title`.
- the 3^rd^ can be send with a **JSON** body, in **multipart** or **URL encoded** it makes no difference and only give clients a choice over the technology to use. If not renamed, the variable will be given to the handler with the name `content`.
