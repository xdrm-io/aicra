# aicra: all-in-config REST API



**Aicra** is a self-working framework coded in *Go* that allows anyone to create a fully featured REST API. It features :

- type checking

- authentication management

- rich argument management (url slash-separated, urlencoded, form-data, json)

- nested routes

- request and response interfaces

- a project builder (cf. *aicra*)

- Optional arguments and default values

- ...



It is based over some of the following concepts.

| concept | explanation |
|---|---|
| meaningful defaults | Defaults and default values work without further understanding |
| config-driven | Avoid information duplication. Automate anything that can be without losing control. Have *one* configuration that summarizes the whole project, its behavior and its automation flow. |




> A working example is available [here](https://git.xdrm.io/example/gfw)



#### 1. Installation

You need a recent machine with `go` installed.

> This package has not been tested under the version **1.10**.



##### (1) Installation

Run the following command to fetch and install the package :

```bash
go get -u git.xdrm.io/go/aicra
```

It should now be available locally and available for your imports.



You should then install the project builder to help you manage your projects, run the following command :

```bash
go install git.xdrm.io/go/aicra/cmd/aicra
```

The executable `aicra` will be placed into your `$GOPATH/bin` folder, if added to your environment PATH it should be available as a standalone command in your terminal. If not, you can simply run `$GOPATH/bin/aicra` to use the command or create a symlink into `/usr/local/bin` or the PATH folder of your choice for less characters to type.



#### 2. Setup a project

To create a project using **aicra**, simply create an empty folder. Then you'll have to create a `manifest.json` file containing your API description. To write your manifest file, follow this [example](https://git.xdrm.io/example/aicra/src/master/manifest.json).



##### (1) Controllers

For each *uri*, you'll have to place your implementation into the local `root` folder following the following naming convention : add `i.go` at the end of the route.

> **Example** - `/path/to/some/uri` will be implemented in the *./root* local folder inside the file : `/path/to/some/urii.go`.

A fully working example is available [here](https://git.xdrm.io/example/aicra).



##### (2) Custom types

If you want to create custom types for the type checker or override built-in types place them inside the `./custom-types` folder. You can check what structure to follow by looking at the [built-in types](https://git.xdrm.io/go/aicra/src/master/checker/default).



#### 3. Build your project

After each controller or custom type edition, you'll have to rebuild the project. This can be achieved through the command-line builder.
Usage is `aicra [options] /path/to/your/project`.

Options:

- `-c controller/path` - overrides the default controllers path ; default is `./root`

- `-t custom/types` - overrides the default custom types path ; default is `./custom-types`






##### changelog

- [x] human-readable json configuration
- [x] nested routes (*i.e. `/user/:id:` and `/user/post/​:id:​`*)
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