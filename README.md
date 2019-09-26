# | aicra |

[![Go version](https://img.shields.io/badge/go_version-1.10.3-blue.svg)](https://golang.org/doc/go1.10)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/git.xdrm.io/go/aicra)](https://goreportcard.com/report/git.xdrm.io/go/aicra)
[![Go doc](https://godoc.org/git.xdrm.io/go/aicra?status.svg)](https://godoc.org/git.xdrm.io/go/aicra)
[![Build Status](https://ci.migration.xdrm.io/buildStatus/icon?job=aicra%2F0.2.0)](.)


**Aicra** is a *configuration-driven* **web framework** written in Go that allows you to create a fully featured REST API.

The whole management is done for you from a configuration file describing your API, you're left with implementing :
- controllers
- optionnally middle-wares (_e.g. authentication, csrf_)
- and optionnally type checkers to check input parameters


The aicra server fulfills the `net/http` [Server interface](https://golang.org/pkg/net/http/#Server).



> A example project is available [here](https://git.xdrm.io/go/tiny-url-ex)


### Table of contents

<!-- toc -->

- [I/ Installation](#i-installation)
- [II/ Development](#ii-development)
  * [1) Main executable](#1-main-executable)
  * [2) API Configuration](#2-api-configuration)
      - [Definition](#definition)
    + [Input Arguments](#input-arguments)
      - [1. Input types](#1-input-types)
      - [2. Global Format](#2-global-format)
- [III/ Change Log](#iii-change-log)

<!-- tocstop -->

### I/ Installation

You need a recent machine with `go` [installed](https://golang.org/doc/install). This package has not been tested under the version **1.10**.


```bash
go get -u git.xdrm.io/go/aicra/cmd/aicra
```

The library should now be available as `git.xdrm.io/go/aicra` in your imports.


### II/ Development


#### 1) Main executable

The main executable will declare and run the aicra server, it might look quite like the code below.

```go
package main
import (
    "log"
    "net/http"

    "git.xdrm.io/go/aicra"
    "git.xdrm.io/go/aicra/typecheck/builtin"
    "git.xdrm.io/go/aicra/api"
)

func main() {

    // 1. build server
    server, err := aicra.New("path/to/your/api/definition.json");
    if err != nil {
        log.Fatalf("Cannot build the aicra server: %v\n", err)
    }

    // 2. add type checkers
    server.Checkers.Add( builtin.NewAny() );
    server.Checkers.Add( builtin.NewString() );
    server.Checkers.Add( builtin.NewFloat64() );

    // 3. bind your implementations
    server.HandleFunc(http.MethodGet, func(req api.Request, res *api.Response){
        // ... process stuff ...
        res.SetError(api.ErrorSuccess());
    })

    // 4. launch server
    log.Fatal( http.ListenAndServer("localhost:8181", server) )
}
```



#### 2) API Configuration

The whole project behavior is described inside a json file (_e.g. usually api.json_) file. For a better understanding of the format, take a look at this working [template](https://git.xdrm.io/go/tiny-url-ex/src/master/api.json). This file defines :

- resource routes and their methods
- every input for each method (called *argument*)
- every output for each method
- scope permissions (list of permissions needed for clients to use which method)
- input policy :
  - type of argument (_i.e. for type checkers_)
  - required/optional
  - default value
  - variable renaming



###### Definition

At the root of the json file are available 5 field names :

1. `GET` - to define what to do when receiving a request with a GET HTTP method at the root URI
2. `POST` - to define what to do when receiving a request with a POST HTTP method at the root URI
3. `PUT` - to define what to do when receiving a request with a PUT HTTP method at the root URI
4. `DELETE` - to define what to do when receiving a request with a DELETE HTTP method at the root URI
5. `/` - to define children URIs ; each will have the same available fields



For each method you will have to create fields described in the table above.

| field path | description                                                  | example                                                      |
| ---------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| `info`     | A short human-readable description of what the method does   | `create a new user`                                          |
| `scope`    | A 2-dimensional array of permissions. The first dimension can be translated to a **or** operator, the second dimension as a **and**. It allows you to combine permissions in complex ways. | `[["A", "B"], ["C", "D"]]` can be translated to : this method needs users to have permissions (A **and** B) **or** (C **and** D) |
| `in`       | The list of arguments that the clients will have to provide. See [here](#input-arguments) for details. |                                                              |
| `out`      | The list of output data that will be returned by your controllers. It has the same syntax as the `in` field but is only use for readability purpose and documentation. |                                                              |


##### Input Arguments

###### 1. Input types

Input arguments defines what data from the HTTP request the method needs. Aicra is able to extract 3 types of data :

- **URI** - Slash-separated strings right after the resource URI. For instance, if your controller is bound to the `/user` URI, you can use the *URI slot* right after to send the user ID ; Now a client can send requests to the URI `/user/:id` where `:id` is a number sent by the client. This kind of input cannot be extracted by name, but rather by index in the URL (_begins at 0_).
- **Query** - data formatted at the end of the URL following the standard [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **URL encoded** - data send inside the body of the request but following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **Multipart** - data send inside the body of the request with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format is not very lightweight but allows you to receive data as well as files.
- **JSON** - data send inside the body as a json object ; each key being a variable name, each value its content. Note that the HTTP header '**Content-Type**' must be set to `application/json` for the API to use it.



###### 2. Global Format

The `in` field in each method contains as list of arguments where the key is the argument name, and the value defines how to manage the variable.

> Variable names must be <u>prefixed</u> when requesting **URI** or **Query** input types.
>
> - The first **URI** data has to be named `URL#0`, the second one `URL#1` and so on...
> - The variable named `somevar` in the **Query** has to be named `GET@somvar` in the configuration.



**Example**

In this example we want 3 arguments :

- the 1^st^ one is send at the end of the URI and is a number compliant with the `int` type checker (else the controller will not be run). It is renamed `uri-param`, this new name will be sent to the controller.
- the 2^nd^ one is send in the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `int` or not given at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `get-param`.
- the 3^rd^ can be send with a **JSON** body, in **multipart** or **URL encoded** it makes no difference and only give clients a choice over the technology to use. If not renamed, the variable will be given to the controller with the name `multipart-var`.

```json
"in": {
    // arg 1
    "URL#0": {
        "info": "some integer in the URI",
        "type": "int",
        "name": "uri-param"
    },
    // arg 2
    "GET@get-var": {
        "info": "some Query OPTIONAL variable",
        "type": "?int",
        "name": "get-param"
    },
    // arg 3
    "multipart-var": { /* ... */ }
}
```



### III/ Change Log

- [x] human-readable json configuration
- [x] nested routes (*i.e. `/user/:id:` and `/user/post/:id:`*)
- [ ] nested URL arguments (*i.e. `/user/:id:` and `/user/:id:/post/​:id:​`*)
- [x] useful http methods: GET, POST, PUT, DELETE
- [x] manage URL, query and body arguments:
  - [x] multipart/form-data (variables and file uploads)
  - [x] application/x-www-form-urlencoded
  - [x] application/json
- [x] required vs. optional parameters with a default value
- [x] parameter renaming
- [x] generic type check (*i.e. implement custom types alongside built-in ones*)
- [ ] built-in types
  - [x] `any` - wildcard matching all values
  - [x] `int` - see go types
  - [x] `uint` - see go types
  - [x] `float` - see go types
  - [x] `string` - any text
  - [x] `string(min, max)` - any string with a length between `min` and `max`
  - [ ] `[a]` - array containing **only** elements matching `a` type
  - [ ] `[a:b]` - map containing **only** keys of type `a` and values of type `b` (*a or b can be ommited*)
- [x] generic controllers implementation (shared objects)
- [x] response interface
- [ ] log bound resources when building the aicra server
- [ ] fail on check for unimplemented resources at server boot.
- [ ] fail on check for unavailable types in api.json at server boot.
