<p align="center">
  <a href="https://github.com/xdrm-io/aicra">
    <img src="https://github.com/xdrm-io/aicra/raw/0.4.0/readme.assets/loGO.png" alt="aicra loGO" width="200" height="200">
  </a>
</p>

<h3 align="center">aicra</h3>

<p align="center">
  Fast, intuitive, and powerful configuration-driven engine for faster and easier <em>REST</em> development.
</p>

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![GO version](https://img.shields.io/badge/GO_version-1.16-blue.svg)](https://GOlang.org/doc/GO1.16) [![GO doc](https://pkg.GO.dev/badge/github.com/xdrm-io/aicra)](https://pkg.GO.dev/github.com/xdrm-io/aicra) [![GO Report Card](https://GOreportcard.com/badge/github.com/xdrm-io/aicra)](https://GOreportcard.com/report/github.com/xdrm-io/aicra) [![Build status](https://github.com/xdrm-io/aicra/actions/workflows/GO.yml/badge.svg)](https://github.com/xdrm-io/aicra/actions/workflows/GO.yml) [![Coverage](https://codecov.io/gh/xdrm-io/aicra/branch/0.4.0/graph/badge.svg?token=HDIMZ0MKXW)](https://codecov.io/gh/xdrm-io/aicra)


# Presentation

`aicra` is a lightweight and idiomatic configuration-driven engine for building REST services. It's especially GOod at helping you write large APIs that remain maintainable as your project grows.

The focus of the project is to allow you to build a fully-featured REST API in an elegant, comfortable and inexpensive way. This is achieved by using a single configuration file to drive the server. This one file describes your entire API: methods, uris, input data, expected output, permissions, etc.

Repetitive tasks are automated by `aicra` based on your configuration, you're left with implementing your endpoints (_usually business logic_).


# Table of contents

- [Presentation](#presentation)
- [Table of contents](#table-of-contents)
- [Installation](#installation)
- [What's automated](#whats-automated)
- [API Documentation](#api-documentation)
- [Getting started](#getting-started)
- [Configuration](#configuration)
    - [Endpoints](#endpoints)
  - [Contextual Permissions](#contextual-permissions)
  - [Parameters](#parameters)
    - [Input extraction](#input-extraction)
    - [Mandatory vs. Optional](#mandatory-vs-optional)
    - [Renaming](#renaming)
    - [Input validators](#input-validators)
    - [Output types](#output-types)
- [Writing endpoint handlers](#writing-endpoint-handlers)
  - [Function signature](#function-signature)
  - [Response formatting](#response-formatting)
- [Example endpoint](#example-endpoint)
  - [Configuration](#configuration-1)
  - [Code](#code)
- [Coming next](#coming-next)


# Installation

To use the aicra package, you need to have GO installed.
> not tested under GO 1.14

1. add aicra to your project
```bash
$ GO get -u github.com/xdrm-io/aicra
```
2. Import in your code
```GO
import "github.com/xdrm-io/aicra"
```


# What's automated

As the configuration file is here to make your life easier, let's take a quick look at what you do not have to do ; or in other words, what does `aicra` automates for you.

HTTP requests and responses are automatically processed.

Requests are only accepted when they meet the permissions you have defined. Otherwise, the request is automatically rejected with an error.

Request data is automatically validated and extracted before it reaches your code. Missing or invalid data results in an automatic error response.

Aicra injects input data into your endpoints and formats the output data back to an http response.

Any error in the configuration or your code is spotted before the server starts and accepts incoming requests. Only when the server is valid (the configuration and your endpoints), it starts listening for incoming requests. Moreover, errors give you enough context to pinpoint and solve the issue effortlessly. There will be no surprise at "runtime" !

You will get errors for:
- invalid configuration syntax
- handler signature not matching the configuration
- configuration endpoint with no handler
- handler matching no endpoint

The same applies if your configuration is invalid:
- unknown HTTP method
- invalid uri
- uri collision between 2 services
- missing fields
- unknown data type
- input name collision
- etc.


# API Documentation

The base idea behind aicra is to avoid the requirement for tooling in addition to the configuration. For OpenAPI, Swagger provides an editor with validation, documentation generation. I strongly believe that this is not required with aicra, the configuration file has been designed to be as descriptive and readable as possible.

It avoids having multiple sources of truth, where your documentation can be outdated. With aicra the same file is used to drive your API server and document your API for other team members as the configuration is versioned alongside your code.


# Getting started

> Other examples are available in the [examples folder](./examples/).

Example `main()` to launch your aicra server.
```GO
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/xdrm-io/aicra"
    "github.com/xdrm-io/aicra/api"
    "github.com/xdrm-io/aicra/validator/builtin"
)

const configFile = "api.json"

func main() {
    builder := &aicra.Builder{}

    // add input validators
    builder.Input(validator.BoolDataType{})
    builder.Input(validator.StringDataType{})

    // add output types
    builder.Output("string", "")
    builder.Output("user", UserStruct{})
    builder.Output("users", []UserStruct{})

    // load your configuration
    config, err := os.Open(configFile)
    if err != nil {
        log.Fatalf("cannot open config: %s", err)
    }
    err = builder.Setup(config)
    config.Close()
    if err != nil {
        log.Fatalf("invalid config: %s", err)
    }

    // add http middlewares (logger, cors)
    builder.With(func(next http.Handler) http.Handler{ /* ... */ })

    // add contextual middlewares (authentication)
    builder.WithContext(func(next http.Handler) http.Handler{ /* ... */ })

    // bind your endpoints to your functions
    err = aicra.Bind(builder, http.MethodGet, "/user/{id}", getUserById)
    if err != nil {
        log.Fatalf("cannot bind: %s", err)
    }

    // build your api
    handler, err := builder.Build()
    if err != nil {
        log.Fatalf("cannot build: %s", err)
    }
    http.ListenAndServe("localhost:8080", handler)
}
```

For HTTPS, you can configure your own [`http.Server`](https://pkg.GO.dev/net/http#Server) :

```GO
server := &http.Server{
	Addr:      "localhost:8080",
	TLSConfig: &tls.Config{},
	// ...
	Handler: handler, // aicra handler
}

server.ListenAndServeTLS("server.crt", "server.key")
```



# Configuration

The configuration uses the `json` syntax.

> Quick note if you thought: "I don't like JSON, I would have preferred yaml, or even xml !"
>
> I've had a hard time deciding and testing different formats including yaml and xml.
> But as it describes our entire api and is crucial for our server to keep working over updates; xml would have been too verbose with growth and yaml on the other side would have been too difficult to read. Json sits in the right spot for this.

Let's take a quick look at the configuration format !

> If you don't like boring explanations and prefer a working example, take a look [here](https://github.com/xdrm-io/articles-api/blob/main/api/definition.json)

The configuration file consists of a list of endpoints.

### Endpoints

The configuration file defines a list of endpoints. Each one is defined by:
- `method` an HTTP method
- `path` an URI pattern (can contain variables)
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
    },
    // ...other endpoints
]
```

The `scope` is a 2-dimensional list of permissions. The first list means **or**, the second means **and**, it allows for complex permission combinations. The example above can be translated to: this method requires users to have permissions (author **and** reader) **or** (admin)


## Contextual Permissions

The `scope` attribute allows to define any combination of permissions, but it lacks context. For instance, in your articles API, an `author` permission protects the modification and deletion of articles. It is your code's responsibility to check that the `author` is the right one according to the requested article.

Aicra provides a way to contextualize permissions. It moves this logic from the code to the configuration when required.

When writing your scopes, you can use the `[Var]` syntax to refer to the `path` variable named `Var`. For each request, the scope automatically replaces `[Var]` with `[XXX]`, `XXX` being the value of the `Var` parameter. The `name` field is used for the variable name (cf. [Rename]().

> It is limited to URI arguments for security reasons.
>
> Allowing GET or body variables in the scope means that an unauthorized party could overload the server with large requests. We must check authentication first in an inexpensive way before extracting its content.


<details>
<summary>
Example
</summary>

In this example we only want the user to update its own information.

We assume that the list of permissions in the request's context is `user[123]` for the user with an id of `123`.

```json
[
    {
        "method": "PUT",
        "path": "/user/{id}/info",
        "scope": [["user[UserID]"], ["admin"]],
        "info": "updates user information ; only authorized for the user itself or the administrator.",
        "in": {
	        "{id}": { "name": "UserID",  "type": "uint", "info": "id of the user to udpate" }
        },
        "out": {}
    }
]
```

- user 456 requests `PUT /user/123/info` -> forbidden
- user 123 requests `PUT /user/123/info` -> accepted


## Parameters

Input and output parameters share the same format, consisting of:
- `info` a short description of what it is
- `type` its data type (_c.f. validation_)
- `?` whether an input parameter is mandatory. It does not work with output parameters.
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

</details>
<br>

### Input extraction

The format of the key for input arguments defines where it comes from:
1. `{param}` is an URI parameter that is extracted from the `"path"`
2. `GET@param` is an URL parameter that is extracted from the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
3. `param` is a body parameter extracted according to the Content-Type.

Body parameters are extracted based on the `Content-Type` header. Supported types are:
- `application/x-www-form-urlencoded` - data send in the body following the [HTTP Query](https://tools.ietf.org/html/rfc3986#section-3.4) syntax.
- `multipart/form-data` - data send in the body with a dedicated [format](https://tools.ietf.org/html/rfc2388#section-3). This format can be quite heavy but allows to transmit data as well as files.
- `application/json` - data sent in the body as a json object

<details>
<summary>Example</summary>

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
            "Content":   { "info": "...", "type": "string"                   }
        },
        "out": {
            "Title":   { "info": "updated article title",   "type": "string" },
            "Content": { "info": "updated article content", "type": "string" }
        }
    }
]
```

In the example above, it reads:
1. `{id}` is extracted from the end of the URI and is a number compliant with the `int` type checker. It is renamed `ID`, this new name will be used by the handler in GO code.
2. `GET@title` is extracted from the query (_e.g. [http://host/uri?get-var=value](http://host/uri?get-var=value)_). It must be a valid `string` or not provided at all (the `?` at the beginning of the type tells that the argument is **optional**) ; it will be named `Title`.
3. `Content` can be extracted from json, multipart or url-encoded data; it makes no difference and only give clients a choice over the technology to use. It is not renamed, the variable will pass to the handler with its original name `Content`.
</details>
<br>

### Mandatory vs. Optional

If you want to make an input parameter optional, prefix its type with a question mark, by default all parameters are mandatory.

When a parameter is optional, the attribute of the GO struct must be a pointer.


### Renaming

Renaming with the field `"name"` is mandatory for:
- URI parameters, the `{var}` syntax
- get parameters, the `GET@var` syntax
- body parameters that do not start with an uppercase letter or contain invalid characters for GO variables

These names are the same as input or output parameters in your code, they must begin with an uppercase letter in order to be exported and valid GO.


### Input validators

Every input type must match one of the input validators registered with [`Builder.Input()`](https://pkg.go.dev/github.com/xdrm-io/aicra#Builder.Input). Aicra provides [built-in validators](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/validator), you can add your own according to your needs. Validators must implement the [`validator.Type`](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/validator#Type) interface.


<details>
<summary>Example validator for any number</summary>

```go
type NumberType struct{}

// GoType returns a float64 as any number can be converted to float64
func (NumberType) GoType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

// Validator for any kind of number value
func (NumberType) Validator(typename string, avail ...validator.Type) ValidateFunc {
	// ignore other type names from the configuration
	if typename != "number" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		switch cast := value.(type) {
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
			return float64(cast), true
		case []byte, string:
			// serialized string -> try to convert to float
			num, err := strconv.ParseFloat(string(cast), 64)
			return float64(num), err == nil
		default:
			return 0, false
		}
	}
}

// main.go
builder.Input(NumberType{})
```

</details>
<br>

The `Validator()` method of the interface seems a bit complicated, this is to allow complex types such as arrays or maps.

The `typename` argument allows to create a dynamic type such as a `varchar` type that can have parameters, i.e. `varchar(123)`. There is an example of such a validator with the [built-in string type](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/validator#StringType).

The `avail` argument allows to build aggregation types, such as arrays of other existing types. The `avail` argument contains all validators of the aicra server.

<details>
<summary>Example array meta type</summary>

> This does not work and has not been tested, but the idea is here.

```go
func (ArrayType) Validator(typename string, avail ...validator.Type) ValidateFunc {
	// matches: []string, []int, []user, ...
	if !strings.HasPrefix(typename, "[]") {
		return nil
	}
	// extracts: string, int, user, ...
	itemTypename := strings.TrimPrefix(typename, "[]")

	// find validator for the type after [] in the typename
	var itemValidator validator.Type
	for _, other := range avail {
		itemValidator = other.Validator(itemTypename, avail)
		if itemValidator != nil { // item validator found
			break;
		}
	}

	// configuration error: validator with the items typename not found
	if itemValidator == nil {
		return nil
	}

	return func(value interface{}) (interface{}, bool) {
		slice, isSlice := value.([]any)
		if !isSlice {
			return []any{}, false
		}

		// validate every item
		for _, item := range slice {
			if _, ok := itemValidator(item); !ok {
				return []any{}, false
			}
		}
		return slice, true
	}
}

// main.go
builder.Input(ArrayType{})
```
</details>
<br>

### Output types

Every output type must match one of the output types registered with [`Builder.Output()`](https://pkg.go.dev/github.com/xdrm-io/aicra#Builder.Output).

No validation is required, you simply have to associate a type name with its GO type.

```go
builder.Output("string", "")           // string
builder.Output("byte",   uint8(0))     // uint8
builder.Output("user",   UserStruct{}) // your custom struct UserStruct
```

> The [Output()](https://pkg.go.dev/github.com/xdrm-io/aicra#Builder.Output) method uses reflection to get the type of the second argument.


# Writing endpoint handlers

Besides your main package where you launch your server, you will need to create a handler for each endpoint defined in your configuration file.


## Function signature

Handler's function signature is defined by the configuration of the endpoint it implements.

Every handler function must feature at least:
- a first input argument of type `context.Context`
- a last output argument of type `error`

Request and/or response struct must be added when defined in the configuration. Here are some basic examples.

<table>
    <thead>
        <tr>
            <td>service configuration (json)</td>
            <td>service handler (go)</td>
        </tr>
    </thead>
    <tbody>
<tr>
<td>

No input with no output
```json
[
    {
        "method": "GET",
        "path": "/users",
        "scope": [],
        "info": "lists all users",
        "in": {},
        "out": {}
    }
]
```

</td>
<td>

```go
func serviceHandler(ctx context.Context) error {
    return nil
}
```

</td>
</tr>
<tr>
<td>

Input with no output
```json
[
    {
        "method": "PUT",
        "path": "/user/{id}",
        "scope": [],
        "info": "updates an existing user",
        "in": {
            "{id}": {
                "name": "ID",
                "type": "uint",
                "info": "target user uid"
            },
            "firstname": {
                "name": "Firstname",
                "type": "?string",
                "info": "new firstname"
            },
            "lastname": {
                "name": "Lastname",
                "type": "?string",
                "info": "new lastname"
            }
        },
        "out": {}
    }
]
```

</td>
<td>

Note: optional input arguments are pointers.
```go
type request {
    ID uint
    Firstname *string
    Lastname *string
}
func serviceHandler(ctx context.Context, req request) error {
    return nil
}
```

</td>
</tr>
<tr>
<td>

No input with output
```json
[
    {
        "method": "GET",
        "path": "/users",
        "scope": [],
        "info": "returns all existing users",
        "in": {},
        "out": {
            "users": {
                "name": "Users",
                "type": "[]User",
                "info": "list of existing users"
            }
        }
    }
]
```

</td>
<td>

```go
type response {
    Users []User
}
func serviceHandler(ctx context.Context) (*response, error) {
    return &response{Users: []User{}}, nil
}
```

</td>
</tr>
<tr>
<td>

Input with output
```json
[
    {
        "method": "PUT",
        "path": "/user/{id}",
        "scope": [],
        "info": "updates an existing user",
        "in": {
            "{id}": {
                "name": "ID",
                "type": "uint",
                "info": "target user uid"
            },
            "firstname": {
                "name": "Firstname",
                "type": "?string",
                "info": "new firstname"
            },
            "lastname": {
                "name": "Lastname",
                "type": "?string",
                "info": "new lastname"
            }
        },
        "out": {
            "user": {
                "name": "User",
                "type": "User",
                "info": "updated user info",
            }
        }
    }
]
```

</td>
<td>

```go
type request {
    ID uint
    Firstname *string
    Lastname *string
}
type response {
    User User
}
func serviceHandler(ctx context.Context, req request) (*response, error) {
    return &response{User: User{}}, nil
}
```

</td>
</tr>
    </tbody>
</table>

If your handler signature does not exactly match the configuration, the server will print out the error and won't start.

## Response formatting

Example parameters configuration :
```json
{
	"in": {
		"input1": { "name": "Input1", "type": "int",     "info": "..." },
		"input2": { "name": "Input2", "type": "?string", "info": "..." }
	},
	"out": {
		"output1": { "name": "Output1", "type": "string", "info": "..." },
		"output2": { "name": "Output2", "type": "bool",   "info": "..." }
	}
}
```
```GO
type req struct{
    Input1 int
    Input2 *string
}
type res struct{
    Output1 string
    Output2 bool
}

func myEndpoint(ctx context.Context, r req) (*res, error) {
    if err := fetchData(req.Input1); err != nil {
        return nil, api.ErrFailure // built-in error
    }
    if req.Input2 != nil {
        if err := fetchData(req.Input2); err != nil {
            return nil, api.Error(404, err) // custom error
        }
    }
    return &res{Output1: "out1", Output2: true}, nil
}
```

The [`api.Err`](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/api#Err) type automatically maps to HTTP status codes and error descriptions that will be sent to the client as json. This way, clients can manage the same format for every response:
```http
HTTP/1.1 404 OK
Content-Type: application/json

{"status":"not found"}
```

By default, responses are formatted using the [DefaultResponder](https://pkg.go.dev/github.com/xdrm-io/aicra#DefaultResponder). The way to format responses can be overwritten with [Builder.RespondWith()](https://pkg.GO.dev/github.com/xdrm-io/aicra#Builder.RespondWith).

Aicra provides [built-in api.Err](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/api#pkg-constants) errors, you can create your own constants or wrap standard errors with the [`api.Error()`](https://pkg.go.dev/github.com/xdrm-io/aicra@v0.4.11/api#Error) method.


# Example endpoint

In this example we will use a endpoint to update an existing article from its id. The optional new title is provided in the URL and the content is provided in the body (not optional).

Some valid HTTP requests :

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


## Configuration

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

## Code

```go
type req struct {
	ID uint
	Title *string
	Content string
}

type res struct {
	Article ArticleStruct
}

func endpoint(ctx context.Context, r req) (*res, error) {
	article, err := db.GetArticleByID(r.ID)
	if err != nil {
		return nil, api.ErrNotFound
	}

	// update the article
	article.Content = r.Content
	if r.Title != nil {
		article.Title = *r.Title
	}

	if err := db.Save(article) ; err != nil {
		return nil, api.ErrUpdate
	}
	return &res{Article: article}, nil
}
```



# Coming next
- [ ] support for PATCH or other custom http methods. It might be interesting to generate the list of allowed methods from the configuration. A check against available http methods as a failsafe might be required.
	- it might be interesting to generate the list of allowed methods from the configuration
- [ ] Consider code generation to avoid using `reflect` that has a big impact on performance as it is used for every incoming request. Some big issues appear with code generation, to be designed properly.
