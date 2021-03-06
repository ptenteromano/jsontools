package jsontools

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Tools struct {
	MaxFileSize int
}

// JSONResponse is the type used for sending JSON around
type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Port    string `json:"port,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Tries to read the body of a request and converts it into JSON
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte

	if t.MaxFileSize > 0 {
		maxBytes = t.MaxFileSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

// Takes a response status code and arbitrary data and writes a json response to the client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil
}

// takes an error, and optionally a response status code, and generates and sends json
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	payload := JsonResponse{
		Error:   true,
		Message: err.Error(),
	}

	return t.WriteJSON(w, statusCode, payload)
}
