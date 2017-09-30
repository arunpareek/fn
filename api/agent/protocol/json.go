package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// This is sent into the function
// All HTTP request headers should be set in env
type JSONIO struct {
	Headers    http.Header `json:"headers,omitempty"`
	Body       string      `json:"body"`
	StatusCode int         `json:"status_code,omitempty"`
}

// JSONProtocol converts stdin/stdout streams from HTTP into JSON format.
type JSONProtocol struct {
	in  io.Writer
	out io.Reader
}

func (p *JSONProtocol) IsStreamable() bool {
	return true
}

func (h *JSONProtocol) Dispatch(w io.Writer, req *http.Request) error {
	_, err := io.WriteString(h.in, `{`)
	if err != nil {
		// this shouldn't happen
		return respondWithError(
			w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
	}
	if req.Body != nil {
		_, err := io.WriteString(h.in, `"body":"`)
		if err != nil {
			// this shouldn't happen
			return respondWithError(
				w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
		}
		_, err = io.CopyN(h.in, req.Body, req.ContentLength)
		if err != nil {
			// this shouldn't happen
			return respondWithError(
				w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
		}
		_, err = io.WriteString(h.in, `",`)
		if err != nil {
			// this shouldn't happen
			return respondWithError(
				w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
		}
		defer req.Body.Close()
	}
	_, err = io.WriteString(h.in, `"headers:"`)
	if err != nil {
		// this shouldn't happen
		return respondWithError(
			w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
	}
	err = json.NewEncoder(h.in).Encode(req.Header)
	if err != nil {
		// this shouldn't happen
		return respondWithError(
			w, fmt.Errorf("error marshalling JSONInput: %s", err.Error()))
	}
	_, err = io.WriteString(h.in, `"}`)
	if err != nil {
		// this shouldn't happen
		return respondWithError(
			w, fmt.Errorf("error reader JSON object from request body: %s", err.Error()))
	}

	jout := new(JSONIO)
	dec := json.NewDecoder(h.out)
	if err := dec.Decode(jout); err != nil {
		return respondWithError(
			w, fmt.Errorf("unable to decode JSON response object: %s", err.Error()))
	}
	if rw, ok := w.(http.ResponseWriter); ok {
		// this has to be done for pulling out:
		// - status code
		// - body
		rw.WriteHeader(jout.StatusCode)
		_, err = rw.Write([]byte(jout.Body)) // TODO timeout
		if err != nil {
			return respondWithError(
				w, fmt.Errorf("unable to write JSON response object: %s", err.Error()))
		}
	} else {
		// logs can just copy the full thing in there, headers and all.
		err = json.NewEncoder(w).Encode(jout)
		if err != nil {
			return respondWithError(
				w, fmt.Errorf("error writing function response: %s", err.Error()))
		}
	}
	return nil
}

func respondWithError(w io.Writer, err error) error {
	errMsg := []byte(err.Error())
	statusCode := http.StatusInternalServerError
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.WriteHeader(statusCode)
		rw.Write(errMsg)
	} else {
		// logs can just copy the full thing in there, headers and all.
		w.Write(errMsg)
	}

	return err
}
