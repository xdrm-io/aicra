# | aicra |

[![Go version](https://img.shields.io/badge/go_version-1.10.3-blue.svg)](https://golang.org/doc/go1.10)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/git.xdrm.io/go/aicra)](https://goreportcard.com/report/git.xdrm.io/go/aicra)
[![Go doc](https://godoc.org/git.xdrm.io/go/aicra?status.svg)](https://godoc.org/git.xdrm.io/go/aicra)
[![Build Status](https://drone.xdrm.io/api/badges/go/aicra/status.svg)](https://drone.xdrm.io/go/aicra)


**Aicra** is a *configuration-driven* **web framework** written in Go that allows you to create a fully featured REST API.

The whole management is done for you from a configuration file describing your API, you're left with implementing :
- handlers
- optionnally middle-wares (_e.g. authentication, csrf_)
- and optionnally your custom type checkers to check input parameters


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

Your main executable will declare and run the aicra server, it might look quite like the code below.

```go
package main

import (
		"log"
		"net/http"

		"git.xdrm.io/go/aicra"
		"git.xdrm.io/go/aicra/datatype"
		"git.xdrm.io/go/aicra/datatype/builtin"
)

func main() {

	// 1. select your datatypes (builtin, custom)
	var dtypes []datatype.T
	dtypes = append(dtypes, builtin.AnyDataType{})
	dtypes = append(dtypes, builtin.BoolDataType{})
	dtypes = append(dtypes, builtin.UintDataType{})
	dtypes = append(dtypes, builtin.StringDataType{})

	// 2. create the server from the configuration file
	server, err := aicra.New("path/to/your/api/definition.json", dtypes...)
	if err != nil {
		log.Fatalf("cannot built aicra server: %s\n", err)
	}

	// 3. bind your implementations
	server.HandleFunc(http.MethodGet, "/path", func(req api.Request, res *api.Response){
		// ... process stuff ...
		res.SetError(api.ErrorSuccess());
	})

	// 4. extract to http server
	httpServer, err := server.ToHTTPServer()
	if err != nil {
		log.Fatalf("cannot get to http server: %s", err)
	}

	// 4. launch server
	log.Fatal( http.ListenAndServe("localhost:8080", server) )
}
```



#### 2) API Configuration

The whole project behavior is described inside a json file (_e.g. usually api.json_). For a better understanding of the format, take a look at this working [template](https://git.xdrm.io/go/tiny-url-ex/src/master/api.json). This file defines :

- routes and their methods
- every input for each method (called *argument*)
- every output for each method
- scope permissions (list of permissions needed by clients)
- input policy :
	- type of argument (_i.e. for data types_)
	- required/optional
	- variable renaming



###### Definition

The root of the json file must be an array containing your requests definitions.

For each, you will have to create fields described in the table above.

| field path | description                                                  | example                                                      |
| ---------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| `info`     | A short human-readable description of what the method does   | `create a new user`                                          |
| `scope`    | A 2-dimensional array of permissions. The first dimension can be translated to a **or** operator, the second dimension as a **and**. It allows you to combine permissions in complex ways. | `[["A", "B"], ["C", "D"]]` can be translated to : this method needs users to have permissions (A **and** B) **or** (C **and** D) |
| `in`       | The list of arguments that the clients will have to provide. See [here](#input-arguments) for details. |                                                              |
| `out`      | The list of output data that will be returned by your controllers. It has the same syntax as the `in` field but is only use for readability purpose and documentation. |                                                              |


##### Input Arguments

###### 1. Input types

Input arguments defines what data from the HTTP request the method needs. Aicra is able to extract 3 types of data :

- **URI** - Curly Braces enclosed strings inside the request path. For instance, if your controller is bound to the `/user/{id}` URI, you can set the input argument `{id}` matching this uri part.
- **Query** - data formatted at the end of the URL following the standard [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **URL encoded** - data send inside the body of the request but following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- **Multipart** - data send inside the body of the request with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format is not very lightweight but allows you to receive data as well as files.
- **JSON** - data send inside the body as a json object ; each key being a variable name, each value its content. Note that the HTTP header '**Content-Type**' must be set to `application/json` for the API to use it.



###### 2. Global Format

The `in` field in each method contains as list of arguments where the key is the argument name, and the value defines how to manage the variable.

> Variable names from **URI** or **Query** must be named accordingly :
>
> - the **URI** variable `{id}` from your request route must be named `{id}`.
> - the variable `somevar` in the **Query** has to be names `GET@somevar`.

**Example**

In this example we want 3 arguments :

- the 1^st^ one is send at the end of the URI and is a number compliant with the `int` type checker. It is renamed `article_id`, this new name will be sent to the handler.
- the 2^nd^ one is send in the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `string` or not given at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `title`.
- the 3^rd^ can be send with a **JSON** body, in **multipart** or **URL encoded** it makes no difference and only give clients a choice over the technology to use. If not renamed, the variable will be given to the handler with the name `content`.

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



### III/ Change Log

- [x] human-readable json configuration
- [x] nested routes (*i.e. `/user/:id:` and `/user/post/:id:`*)
- [x] nested URL arguments (*i.e. `/user/:id:` and `/user/:id:/post/​:id:​`*)
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
- [x] log bound resources when building the aicra server
- [x] fail on check for unimplemented resources at server boot.
- [x] fail on check for unavailable types in api.json at server boot.
