package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Pet defines the pet model
type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag,omitempty"`
}

// Error defines the error model
type Error struct {
	Code    int32
	Message string
}

var pets = []Pet{}

func writeError(w http.ResponseWriter, status int, message string) {
	resp, _ := json.Marshal(Error{
		Code:    int32(status),
		Message: message,
	})

	w.WriteHeader(status)
	w.Header().Set("Content-type", "application/json")
	w.Write(resp)
}

func listPets(w http.ResponseWriter, r *http.Request) {
	if len(pets) == 0 {
		// if control enters this block, middleware should be the one too respond
		// since this response is not defined in the spec.
		writeError(w, http.StatusNotFound, "no pets found!")
		return
	}

	resp, err := json.Marshal(pets)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	// w.Header().Add("x-next", "it is not a required thing")
	w.Write(resp)
}

func createPet(w http.ResponseWriter, r *http.Request) {
	pet := Pet{}
	err := json.NewDecoder(r.Body).Decode(&pet)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unable to parse request body")
		return
	}

	pets = append(pets, pet)
	// w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(nil)
}

func getPet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id should be an int64")
		return
	}

	for _, pet := range pets {
		if pet.ID == id {
			resp, err := json.Marshal(pet)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			w.Header().Add("Content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return
		}
	}

	writeError(w, http.StatusNotFound, "pet not found")
}

// RegisterRoutes registers route handlers to the passed router
func RegisterRoutes(router *mux.Router) {
	router.Path("/pets").Methods("GET").HandlerFunc(listPets)
	router.Path("/pets").Methods("POST").HandlerFunc(createPet)
	router.Path("/pets/{id:[0-9]+}").Methods("GET").HandlerFunc(getPet)
}
