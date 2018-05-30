# xfw: all-in-one REST API



**xfw** is a self-working API coded in *Go*:

- It allows anyone to create a fully featured REST API
- it can be used in *Go* as a library, or as a proxy to launch external commands (*e.g. php scripts*)

It is based on the *all-in-config* idea, where you only have a configuration file, and your implementation and it works without no further work.



Here's a taste of some features:

-  human-readable json configuration
- nested routes (*i.e. `/user/:id:` vs. `/user/post/​:id:​`*)
- useful http methods: GET, POST, PUT, DELETE
- manage URL, query and body arguments:
  - multipart/form-data
  - application/x-www-form-urlencoded
  - application/json
- required vs. optional parameters with a default value
- parameter renaming
- generic authentication system (*i.e. you can override the built-in*)
- generic type check (*i.e. implement your types alongside built-in types*)