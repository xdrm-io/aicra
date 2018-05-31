# xfw: all-in-one REST API



**xfw** is a self-working API coded in *Go*:

- It allows anyone to create a fully featured REST API
- it can be used in *Go* as a library, or as a proxy to launch external commands (*e.g. php scripts*)

It is based on the *all-in-config* idea, where you only have a configuration file, and your implementation and it works without no further work.




##### changelog

- [x] human-readable json configuration
- [x] nested routes (*i.e. `/user/:id:` and `/user/post/​:id:​`*)
- [ ] nested URL arguments (*i.e. `/user/:id:` and `/user/:id:/post/​:id:​`*)
- [x] useful http methods: GET, POST, PUT, DELETE
- [x] manage URL, query and body arguments:
  - [x] multipart/form-data (variables and file uploads)
  - [x] application/x-www-form-urlencoded
  - [x] application/json
- [x] required vs. optional parameters with a default value
- [x] parameter renaming
- [ ] generic authentication system (*i.e. you can override the built-in one*)
- [x] generic type check (*i.e. implement custom types alongside built-in ones*)
- [ ] built-in types
	- [x] `any` - wildcard matching all values
	- [x] `int` - any number (*e.g. float, int, uint*)
	- [x] `string` - any text
	- [x] `varchar(min, max)` - any string with a length between `min` and `max`
	- [ ] `<a>` - array containing **only** elements matching `a` type
	- [ ] `<a:b>` - map containing **only** keys of type `a` and values of type `b` (*a or b can be ommited*)
- [ ] generic controllers implementation (shared objects)
- [ ] response interface