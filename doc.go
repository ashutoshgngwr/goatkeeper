// Copyright 2019 Ashutosh Gangwar
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package goatkeeper is an HTTP middleware for Golang that validates
HTTP requests and their responses according to types defined
in OpenAPI Specification.

Built using kin-openapi implementation!
https://github.com/getkin/kin-openapi

Usage:
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

With Gorilla mux

	r := mux.NewRouter()
	r.Use(middleware)
	// r.Handle ...

With vanilla Golang

	s := &http.Server{
	  Addr:    ":8080",
	  Handler: middlware(myAPIHandler),
	}

	log.Fatal(s.ListenAndServe())
*/
package goatkeeper
