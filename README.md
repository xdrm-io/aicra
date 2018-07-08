# aicra: all-in-config REST API



**Aicra** is a self-working framework coded in *Go* that allows anyone to create a fully featured REST API. It features type checking, authentication management through middlewares, file upload, rich argument parsing (*i.e. url slash-separated, urlencoded, form-data, json*), nested routes, project builder (*i.e. aicra*), etc.



This framework is based over some of the following concepts.

| concept | explanation |
|---|---|
| meaningful defaults | Defaults and default values work without further understanding |
| config driven | Avoid information duplication. Automate anything that can be without losing control. Have *one* configuration that summarizes the whole project, its behavior and its automation flow. |




> A working example is available [here](https://git.xdrm.io/example/aicra)



#### 1. Installation

You need a recent machine with `go` installed.

> This package has not been tested under the version **1.10**.



##### (1) Download and install the package

```bash
$ go get -u git.xdrm.io/go/aicra
```

It should now be available locally and available for your imports.



##### (2) Compile the command-line builder

You should then compile the project builder to help you manage your projects.

```bash
$ go install git.xdrm.io/go/aicra/cmd/aicra
```



> The executable `aicra` will be placed into your `$GOPATH/bin` folder, if added to your environment PATH it should be available as a standalone command in your terminal. If not, you can simply run `$GOPATH/bin/aicra` to use the command or create a symlink into `/usr/local/bin` or the PATH folder of your choice for less characters to type.



#### 2. Setup a project

The default project structure for **aicra** is as follows :

```
├── main.go          - the entry point
├── manifest.json    - the configuration file
├── middleware       - middleware implementations
├── controller       - controller implementations
└── types            - custom type for the type checker

```

In order for your project to be run, each controller, middleware and type have to be compiled as *plugins* (*i.e. shared objects*). They can then be loaded by the server.



##### (1) Configuration

The whole project behavior is described inside the `manifest.json` file. For a better understanding of the format, take a look at this working [template](https://git.xdrm.io/example/aicra/src/master/manifest.json). This file contains information about :

- resources routes and their methods
- every input for each method (called *argument*)
- every output for each method
- scope permissions
- input policy : type, required/optional, default value, variable renaming, etc.



##### (2) Controllers

For each route, you'll have to place your implementation into the `controller` folder following the naming convention : add `/i.go` at the end of the route.

> Example - `/path/to/some/uri` will be inside `controller/path/to/some/uri/i.go`.

A fully working example is available [here](https://git.xdrm.io/example/aicra).



##### (3) Middlewares

In order for your project to manage authentication, the best solution is to create middlewares, there are programs that updates a *Scope* according to internal or persistent (*i.e.* database) information and the actual http request. They are all run before each request it routed by aicra. The scope are used to match the `scope` field in the configuration file and automatically block non-authed requests. Scopes can also be used for implementation-specific behavior.



Each middleware must be directly inside the `middleware` folder.

> Example - the `1-authentication` middleware will be inside `middleware/1-authentication/main.go`. 

**Note** - middleware execution will be ordered by name. Prefixing your middlewares with their order is a good practice.



##### (4) Custom types

In your configuration you will have to use built-in types (*e.g.* int, any, varchar), but if you want project-specific ones, you can add your own types inside the `type` folder. You can check what structure to follow by looking at the [built-in types](https://git.xdrm.io/go/aicra/src/master/checker/default).



Each type must be inside a unique package directly inside the `type` folder. The package name is arbitrary and does not have to match the name (but it is better if it is explicit), because the `Match()` method already matches the name.



#### 3. Build your project

After each controller, middleware or type edition, you'll have to rebuild the project. This can be achieved through the command-line builder. 

Usage is `aicra [options] /path/to/your/project`.

Options:

- `-c` - overrides the default controllers path ; default is `./controller`
- `-m` - overrides the default middlewares path ; default is `./middleware`
- `-t` - overrides the default custom types path ; default is `./custom-types`



For a project that does not need a different structure, you just have to run this command under your project root

```bash
$ aicra .
```

The output should look like ![that](/storage/git.xdrm.io/GOPATH/src/git.xdrm.io/go/aicra/README.assets/1531039386654.png).

#### 4. Main

The main program is pretty small, it is as followed :

```go
package main

import (
    "log"
    "git.xdrm.io/go/aicra"
)

func main() {

	// 1. init with manifest file
	server, err := aicra.New("manifest.json")
	if err != nil {
		log.Fatalf("cannot load config : %s\n", err)
	}
	
	fmt.Printf("[Server up] 0.0.0.0:4242\n")
	// 2. Launch server
	err = server.Listen(4242)
	if err != nil {
		log.Fatalf("[FAILURE] server failed : %s\n", err)
	}
}
```






##### changelog

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
- [ ] generic authentication system (*i.e. you can override the built-in one*)
- [x] generic type check (*i.e. implement custom types alongside built-in ones*)
- [ ] built-in types
	- [x] `any` - wildcard matching all values
	- [x] `int` - any number (*e.g. float, int, uint*)
	- [x] `string` - any text
	- [x] `varchar(min, max)` - any string with a length between `min` and `max`
	- [ ] `<a>` - array containing **only** elements matching `a` type
	- [ ] `<a:b>` - map containing **only** keys of type `a` and values of type `b` (*a or b can be ommited*)
- [x] generic controllers implementation (shared objects)
- [x] response interface