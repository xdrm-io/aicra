```yaml
module: api
version: 1.2
```

For developers
====

[**I.** Overview](#i-overview)

- [**1** Introduction & features](#1-introduction--features)
- [**2** Basic knowledge](#2-basic-knowledge)

[**II.** Usage](#ii-usage)

- [**1** REST API](#1-rest-api)
- [**2** Internal use (PHP)](#2-internal-use)

[**III.** Configuration](#iii-configuration)

- [**1** Configuration format](#1---configuration-format)
- [**2** Method definition](#2---methoddefinition)
	- [**2.1** Method description](#21---methoddescription)
	- [**2.2** Method permissions](#22---methodpermissions)
	- [**2.3** Method parameters](#23---methodparameters)
		- [parameter name](#parametername)
		- [parameter description](#parameterdescription)
		- [parameter type](#parameterchecker_type)
		- [parameter optional vs. required](#parameteris_optional)
		- [parameter rename](#parameterrename)
		- [parameter default value](#parameterdefault_value)
	- [**2.4** Method output format](#24---methodoutput_format)
	- [**2.5** Method options](#25---methodoptions)

[**IV.** Implementation](#iv-implementation)

- [**1** Permissions : AuthSystem](#1---permissions--authsystem)
- [**2** Core implementation](#2---core-implementation)
	- [Classes](#classes)
	- [Methods](#methods)
	- [Method arguments](#method-arguments)
	- [Return statement](#return-statement)
	- [Before and After scripts](#before-and-after-scripts)
	- [Example](#example)

[**V.** Type Checker](#v-type-checker)

- [**1** Default Types](#1---default-types)
- [**2** Complex Types](#2---complex-types)

[**VI.** Documentation](#vi-documentation)

- [**1** API accessible documentation](#1---api-accessible-documentation)


For clients
====

[**I.** Simple request](#i-simple-request)

- [**1** URL](#1---URL)
- [**2** variables](#2---variables)

[**II.** Usage](#ii-usage)




# **I.** Overview
## **1** Introduction & features

The `api` package (v1.2) allows you to easily create and manage a REST API for your applications.

The aim of this package is to make your life easier working with APIs and internal delegation. The only things you have to do is to implement your controllers, middle-wares and write 2 configuration files, the package will do the rest.

Things you **have** to do :
- write the project configuration file (cf. [project configuration](#iii-project-configuration))
- write the API definition file (cf. [api definition](#iv-api-definition))
- implement your middle-wares to manage authentication, csrf, etc (cf. [AuthSystem](#1---middle-wares))
- implement your controllers (cf. ???)
- implement additional project-specific type checkers (cf. ???)

Things you **don't have** to do :
- check the input variables (cf. [Checker](#v-type-checker))
- multiple permission management
- optional or required input
- Form data type : x-www-urlencoded, multipart, or file upload
- URL variables (slash separated at the end of the route, like standard uri)
- and a lot more ...

## **2** Basic knowledge

The API uses the routes defined in the _api.json_ configuration to export your code (in compiled go or other language). Middle-wares are executed when receiving a new request, then the controllers will be loaded according to the URI.

<u>Example:</u>
* the module `article` contains methods:
	* `GET` to get article data
	* `POST` to post a new article
	* `PUT` to edit an existing article
	* `DELETE` to delete an exisint article

*Note that these are all the available methods handled by aicra*




# **III.** Project Configuration

The documentation consists of directions for how to compile and run the project controllers, middle-wares and type checkers.

## **1** - configuration format

```json
{
	"root": "<project-root>",
    "driver": "<driver>",
    "types": {
        "default": true,
		"folder": "<custom-types-folder>"        
    },
    "controllers": {
        "folder": "<controllers-folder"
    },
    "middlewares": {
        "folder": "<middlewares-folder"
    }
}
```

**Note**: The following fields are optional :

- `types.default` is by default "types"
- 



# **III.** API Definition

The documentation consists of a list of URIs each one can contain several HTTP method specifications.

## **1** - configuration format

The configuration is a set of `uri` paths that can contain up to 4 method types: **POST**, **DELETE**, **PUT**, **GET** or the `/` field to define sub URIs (recursive).

For instance the 4 methods directly inside `/`.`uri1` will be triggered when the calling URI is `/uri1`, the ones directly inside `uri2` will be triggered when calling `/uri1/uri2` and so on..

You can also set full paths if you don't need transitional methods, for instance the path `/`.`uri5`.`/`.`uri6`.`/`.`uri7` will be triggered by the url `/uri5/uri6/uri7`.

**Example:** The example file loaded with the default configuration can be found [here](./../../src/config/api/3.0/modules.json).

```json
{

	"/": {
        "uri1" : {
            "GET":    method.definition,
            "POST":   method.definition,
            "PUT":    method.definition,
            "DELETE": method.definition,

            "/": {
                "uri2": {
                    "GET":    method.definition,
                    "POST":   method.definition,
                    "PUT":    method.definition,
                    "DELETE": method.definition
                }
            }
        },
        "uri5":
        	"/": {
        		"uri6": {
        			"/": {
        				"uri7": {
                            "GET":    method.definition,
                            "POST":   method.definition,
                            "PUT":    method.definition,
                            "DELETE": method.definition
    					}
					}
				}
			}
		}
	}
}
```

**Note**: It is possible to trigger the *root uri* (`/`), so you can set methods directly at the root of the JSON file.

## **2** - method.definition

```json
{
	"info":  method.description,
	"scope": method.permissions,
	"in":    method.parameters,
	"out":  method.output_format
}
```

## **2.1** - method.description

The *description* field must be a **string** containing the human-readable description of what the method does.

## **2.2** - method.permissions

The *permissions* field must be an array. You can manage **OR** and **AND** permission combinations.

- **OR** is applied between each **0-depth** array
- **AND** is applied between each **1-depth** array

For instance the following permission `[ [a,b], [c] ]` means you need the permissions **a** and **b** combined or only the permission **c**.

## **2.3** - method.parameters

The *parameters* field must be an object containing each required or optional parameter needed for the implementation.

```json
"parameter.name": {
	"info":    parameter.description,
	"type":    parameter.checker_type,
	"name":    parameter.rename,
	"default": parameter.default_value,
}
```

#### parameter.name

The *name* field must be a **string** containing variable name that will be asked for the caller.

Note that you can set any string for **body parameters**, but you can also catch :

- **URI** parameters by prefixing your variable name with `URL#` and followed by a number which is the index in the URL (starts with 0). For instance the **first** URI parameter received from `/path/to/controller/some_value` has to be named `URL#0` inside this configuration. It is a good practice to rename these parameters for better access in your code.

- **GET** parameters by prefixing your variable name with `GET@`. For instance the  get parameter **somevar** received from `/host/uri?somevar=some_value` has to be named `GET@somevar` inside this configuration.

#### parameter.description

The *description* field must be a **string** containing the human-readable description of what the parameter is or must be.

#### parameter.checker_type

The *checker_type* field must be a **string** corresponding to a `\api\core\Checker` type that will be checked before calling the implementation.

#### parameter.is_optional

If the parameter is optional and can be ignored by clients, you must prefix _parameter.checker_type_ with a question mark. For instance a number variable will have a type of `number`, if the variable is optional, the type will then be `?number`.

**Note :** it is recommended to add a default value for optional parameters.

#### parameter.rename

The *name* field must be a **string** corresponding to the wanted *variable name* that will be passed to the controller. It is mainly useful for **URI parameters** because their name is not explicit at all.

If omitted, by default, `parameter.name` will be used.

#### parameter.default_value

The *default_value* field must be of compliant to the variable type checker, it will be used only for **optional** parameters when omitted by the client.

By default, each optional parameter will exist and will be set to `null` to the implementation.

## **2.4** - method.output_format

The *output_format* field must have the same format as `method.parameters` but will only be used to generate a documentation or to tell other developers what the method returns if no error occurs. It allows better overview of an API without looking at the code.

<!-- ## **2.5** - method.options

The *options* field must be an **object** containing the available options.

The only option available for now is:

```json
"download": true
```

Your implementation must return 2 fields:

- `body` a string containing the file body
- `headers` an array containing as an associative array file headers

If the API is called with HTTP directly, it will not print the **json** response (only on error), but will instead directly return the created file.

*AJAX:* If called with ajax, you must give the header `HTTP_X_REQUESTED_WITH` set to `XMLHttpRequest`. In that case only, it will return a normal JSON response with the field `link` containing the link to call for downloading the created file. **You must take care of deleting not used files** - there is no such mechanism. -->



# **IV.** Implementation

## **1** - middle-wares

Each middle-ware must implement the [driver.Middleware](https://godoc.org/git.xdrm.io/go/aicra/driver#Middleware) interface.

The `Inspect(http.Request, *[]string)` method gives you the actual http request. And you must edit the string list to add scope elements. After all middle-wares are executed, the final scope is used to check the **method.permission** field to decide whether the controller can be accessed.

### example

For instance here, we check if a token is sent inside the **Authorization** HTTP header. The token if valid defines scope permissions.

```go
package main

import (
	"git.xdrm.io/aicra/driver"
    "net/http"
)

// for the API code to export our middle-ware
func Export() driver.Middleware { return new(TokenManager) }

// mockup type to implement the driver.Middleware interface
type TokenManager interface{}

func (tm TokenManager) Inspect(req http.Request, scope *[]string) {
    token := req.Header.Get("Authorization")
    if !isTokenValid(token) {
        return
    }
    
    scope = append(scope, getTokenPermissions(token)...)
}
```

*Note*: Functions `isTokenValid()` and `getTokenPermissions()` do not exist, it was in order for all to understand the example.


## **2** - controller

Each controller must implement the [driver.Controller](https://godoc.org/git.xdrm.io/go/aicra/driver#Controller) interface.

The `Get(response.Arguments) response.Response` method defines the controller implementation for the `GET` method. Other methods are `Post`, `Put`, and `Delete`. Each controller will be given the arguments according to the <u>api definition</u>.

### example

For instance here, we implement a simple **user** controller.

```go
package main

import (
	"git.xdrm.io/aicra/driver"
	"git.xdrm.io/aicra/response"
	"git.xdrm.io/aicra/err"
)

// for the API code to export our middle-ware
func Export() driver.Controller { return new(UserController) }

// mockup type to implement the driver.Controller interface
type UserController interface{}

func (ctl UserController) Get(args response.Arguments) response.Response {
    res := response.New()
    
    // extract user ID argument
    user_id, ok := args["user_id"].(float64)
    if !ok {
        res.Err = e.Failure
        return *res
    }
    
    // fetch user data
    res.Set("user", getUserData(user_id))
    return res
}

func (ctl UserController) Post(args response.Arguments) response.Response {
    res := response.New()
    
    // extract user ID argument
    username, ok := args["username"].(string)
    password, ok1 := args["password"].(string)
    
    if !ok || !ok1 {
        res.Err = e.Failure
        return *res
    }
    
    // store user data
    res.Set("created", storeUser(username, password))
    return res
}
```

*Note*: Functions `getUserData(int) map[string]string` and `storeUser(string,string) bool` do not exist, it was in order for all to understand the example.




# **V.** Type Checker

Each type checker checks http request extracted values according to the type given in the api definition.

The default types below are available in the `$GOPATH/src/git.xdrm.io/go/aicra/internal/checker/default`, they are loaded by default in any project. You can choose not to load them by settings `types`.`default` to `false` in the project configuration. To add custom types you must implement the [driver.Checker](https://godoc.org/git.xdrm.io/go/aicra/driver#Checker) interface and export it in order for aicra to dynamically load it. 

## **1** - Default types

**Warning :** All received values are extracted using JSON format if possible, else values are considered raw strings.

- `2`, `-3.2` are extracted as floats
- `true` and `false` are stored as booleans
- `null` is stored as null
- `[x, y, z]` is stored as an array containing **x**, **y**, and **z** that can be themselves JSON-decoded if possible.
- `{ "a": x }` is stored as a map containing **x** at the key "a", **x** can be itself JSON-decoded if possible.
- `"some string"` is stored as a string
- `some string` here, the decoding will fail (it is no valid JSON) so it will be stored as a string

|Type|Example|Description|
|---|---|---|
|`any`|`[9,"a"]`, `"a"`|Any data (can be simple or complex)|
|`id`|`10`, `23`|Positive integer number|
|`int`|`-10`, `23`|Any integer number|
|`float`|`-10.2`, `23.5`|Any float|
|`string`|`"Hello!"`|String that can be of any length (even empty)|
|`digest(L)`|`"4612473aa81f93a878..."`|String with a length of  `L`, containing only hexadecimal lowercase characters.|
|`array`|`[]`, `[1, "a"]`|Any array|
|`bool`|`true`, `false`|Boolean|
|`varchar(a,b)`|`"Hello!"`|String with a length between `a` and `b` (included)|

## **2** - Complex types

|Type|Sub-Type|Description|
|---|---|---|
|`array<a>`|`a`|Array containing only entries matching the type `a`|
|`FILE`|_a raw file send in `multipart/form-data`|A raw file sent by `multipart/form-data`|

> **Note:** It is possible to chain `array` type as many as needed.

**Ex.:** `array<array<id>>` - Will only match an array containing arrays that only contains `id` entries.



# **VI.** Documentation

With the *all-in-config* method, we can generate a consistent documentation or other documents from the `/config/modules.json` file.

## **1** - API accessible documentation

You can request the API for information about the current URI by using the `OPTIONS` HTTP method.


====


# **I.** Simple request
## **1** - URL

### format
The `uri` format is as defined: `{base}/{path}/{GET_parameters}`, where
- `{base}` is the server's *API* base uri (ex: `https://example.com/api/v1` or `https://api.exampl.com`)
- `{path}` is the effective path you want to access (ex: `article/author`)
- `{GET_parameters}` is a set of slash-separated values (ex: `val0/val1/val2//val4`)

*Note:* GET parameters are not used as usual (`?var1=val1&var2=val2...`), instead the position in the URL gives them an implicit name which is `URL#`, where `#` is the index in the uri (beginning with 0).

### example 1

If you want to edit an article with the server's REST API, it could be defined as following:

```yaml
http: PUT
path: article/{id_article}
input:
  body: new content of the article
output:
  updated: the updated article data
```

Let's take the example where you want to update the article which id is **23** and set its body to "**blabla new content**"

`HTTP REQUEST`
```
PUT article/23 HTTP/1.0

body=blabla+new+content
```


`HTTP RESPONSE`
```
HTTP/1.0 200 OK
Content-Type: application/json

{
	"error": 0,
	"ErrorDescription": "all right",

	"updated": {
		"id_article": 23,
		"title": "article 23",
		"body":  "blabla new content"
	}

}
```

### example 2

If you want to get a specific article line, the request could be defined as following

```yaml
http: GET
path: article/line/{id_article}/{no_line}
input: -
output:
  articles: the list of matching lines
```

Let's take the example where you want to get **all articles** because `id_article` is set to optional, but you only want the first line of each so you have to give only the second parameter set to `1` (first line).

*Solution:* The position in the `uri` where `id_article` must be, have to be left empty: it will result of 2 slashes (`//`).

`HTTP REQUEST`
```
GET article/line//1 HTTP/1.0
```


`HTTP RESPONSE`
```
HTTP/1.0 200 OK
Content-Type: application/json

{
	"error": 0,
	"ErrorDescription": "all right",

	"articles": [
		{
			"id_article": 23,
			"line": [ 0: "blabla new content"
		},{
			"id_article": 25,
			"line": [ 0: "some article"
		}
	]

}
```