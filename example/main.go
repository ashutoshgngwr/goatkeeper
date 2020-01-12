package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ashutoshgngwr/goatkeeper"
	"github.com/gorilla/mux"
)

func main() {
	file, err := os.Open("./petstore.yaml")
	if err != nil {
		log.Fatalf("unable to open specification file: %s", err.Error())
	}

	spec, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("unable to read specification file: %s", err.Error())
	}

	middleware, err := goatkeeper.NewMiddleware(&goatkeeper.MiddlewareOptions{
		OpenAPISpec:      spec,
		ValidateResponse: true,
		InvalidRequestResponse: &goatkeeper.MiddlewareResponse{
			Status: http.StatusBadRequest,
			Body:   []byte("Goatkeeper: Bad request"),
		},
		InvalidResponseResponse: &goatkeeper.MiddlewareResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("Goatkeeper: Internal server error"),
		},
	})

	if err != nil {
		log.Fatalf("unable to create goatkeeper middlware: %s", err.Error())
	}

	router := mux.NewRouter()
	router.Use(middleware)
	RegisterRoutes(router)

	log.Fatal(http.ListenAndServe(":8080", router))
}
