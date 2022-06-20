package jsontools

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Used to satisfy the interface requirements for http.Client
type RoundTripFunc func(req *http.Request) *http.Response

// Used to satisfy the interface requirements for http.Client
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// Returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

/*
 *  ----- Start of tests testing -----
 */
func Test_ReadJson(t *testing.T) {
	testApp := Tools{
		MaxFileSize: 1048576 * 2,
	}

	// Create a sample json file and add it to the body
	sampleJson := map[string]any{
		"foo": "bar",
	}
	body, _ := json.Marshal(sampleJson)

	// Declare a variable to read the decoded json into
	var decodedJson struct {
		Foo string `json:"foo"`
	}

	// Create a request witht he body
	req, err := http.NewRequest("POST", "/", bytes.NewReader(body))
	if err != nil {
		t.Log("error", err)
	}

	// Create a test response recorder which satisfies the requirments for ResponseWriter
	rr := httptest.NewRecorder()
	defer req.Body.Close()

	// Call ReadJson and check for error
	err = testApp.ReadJSON(rr, req, &decodedJson)
	if err != nil {
		t.Error("failed to decode json: ", err)
	}

	badJson := `
		{ "foo": "bar" }
		{ "alpha": "beta" }
	`

	// Create a request with the bad body
	req, err = http.NewRequest("POST", "/", bytes.NewReader([]byte(badJson)))
	if err != nil {
		t.Log("error", err)
	}

	err = testApp.ReadJSON(rr, req, &decodedJson)
	if err == nil {
		t.Error("did not get an error with bad json")
	}
}

func Test_WriteJson(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	payload := JsonResponse{
		Error:   false,
		Message: "test msg",
	}

	headers := http.Header{}
	headers.Add("FOO", "BAR")
	err := testApp.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write json: %v", err)
	}
}

func Test_ErrorJson(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	err := testApp.ErrorJSON(rr, errors.New("test error"))
	if err != nil {
		t.Error(err)
	}

	var requestPayload JsonResponse

	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&requestPayload)
	if err != nil {
		t.Error("received error when decoding ErrorJson payload:", err)
	}

	if !requestPayload.Error {
		t.Error("error set to false in response from ErrorJson, should be set to true")
	}

	// test with status
	err = testApp.ErrorJSON(rr, errors.New("error with status code"), http.StatusServiceUnavailable)
	if err != nil {
		t.Error(err)
	}
}
