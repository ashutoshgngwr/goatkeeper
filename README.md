# Goatkeeper

[![Build status](https://img.shields.io/github/workflow/status/ashutoshgngwr/goatkeeper/Integration)](https://github.com/ashutoshgngwr/goatkeeper/actions)
[![codecov](https://codecov.io/gh/ashutoshgngwr/goatkeeper/branch/master/graph/badge.svg)](https://codecov.io/gh/ashutoshgngwr/goatkeeper)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/616e6e1d40124408a360a651308ae133)](https://www.codacy.com/manual/ashutoshgngwr/goatkeeper?utm_source=github.com&utm_medium=referral&utm_content=ashutoshgngwr/goatkeeper&utm_campaign=Badge_Grade)
[![Latest release](https://img.shields.io/github/v/release/ashutoshgngwr/goatkeeper.svg)](https://github.com/ashutoshgngwr/goatkeeper/releases)
[![Godoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/ashutoshgngwr/goatkeeper)
[![License](https://img.shields.io/badge/License-Apache%202.0-orange.svg)](https://opensource.org/licenses/Apache-2.0)

GoatKeeper is an HTTP middleware for Golang that validates
HTTP requests and their responses according to types defined
in OpenAPI Specification.

## Goals

- Single source of truth for APIs using OpenAPI specifications
- Ensuring request and response conformity to the specification

## Installation

```shell
$ go get -u github.com/ashutoshgngwr/goatkeeper
```

## Usage

```go
spec, err := ioutil.ReadFile("./openapi-spec.yaml")
if err != nil {
  panic("unable to read openapi spec file")
}

middleware, err := goatkeeper.NewMiddlware(
  &goatkeeper.MiddlewareOptions{
    OpenAPISpec:      spec,
    ValidateResponse: true,
  })

if err != nil {
  panic("unable to initialize goatkeeper middleware")
}
```

### With Gorilla mux

```go
r := mux.NewRouter()
r.Use(middleware)
// r.Handle ...
```

### With vanilla Golang

```go
s := &http.Server{
  Addr:    ":8080",
  Handler: middlware(myAPIHandler),
}

log.Fatal(s.ListenAndServe())
```

## Credits

Built using [kin-openapi](https://github.com/getkin/kin-openapi) implementation!

## License

```
Copyright 2019 Ashutosh Gangwar

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
