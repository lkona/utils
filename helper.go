package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048579 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}
	/*Our first call to dec.Decode(data) decodes the first JSON entry into data ({ "foo": "bar" } in your example here). Then the second call to dec.Decode with the empty struct decodes the second JSON entry into the empty struct ( { "alpha": 2 } in your example). We use an empty struct that isn't tracked by a variable because we don't care what the second JSON entry is, we're just checking to see if there is anything extra to decode. If there wasn't an additional JSON entry then we'd get the io.EOF error from dec.Decode and know the payload only contained a single JSON entry and is valid for our purposes.*/
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have a single JSON value")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
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

func ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return WriteJSON(w, statusCode, payload)
}
