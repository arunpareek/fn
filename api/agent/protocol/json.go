package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JSONInput is what's sent into the function
// All HTTP request headers should be set in env
type JSONInput struct {
	RequestURL string `json:"request_url"`
	CallID     string `json:"call_id"`
	Method     string `json:"method"`
	Body       string `json:"body"`
}

// JSONOutput function must return this format
// StatusCode value must be a HTTP status code
type JSONOutput struct {
	StatusCode int    `json:"status"`
	Body       string `json:"body"`
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
	reqURL := req.Header.Get("REQUEST_URL")
	method := req.Header.Get("METHOD")
	callID := req.Header.Get("CALL_ID")

	// TODO content-length or chunked encoding
	var body bytes.Buffer
	if req.Body != nil {
		var dest io.Writer = &body

		// TODO copy w/ ctx and check err
		io.Copy(dest, req.Body)
	}

	// convert to JSON func format
	jin := &JSONInput{
		RequestURL: reqURL,
		Method:     method,
		CallID:     callID,
		Body:       body.String(),
	}
	b, err := json.Marshal(jin)
	if err != nil {
		// this shouldn't happen
		return fmt.Errorf("error marshalling JSONInput: %v", err)
	}
	h.in.Write(b)

	// TODO: put max size on how big the response can be so we don't blow up
	jout := &JSONOutput{}
	dec := json.NewDecoder(h.out)
	if err := dec.Decode(jout); err != nil {
		// TODO: how do we get an error back to the client??
		return fmt.Errorf("error unmarshalling JSONOutput: %v", err)
	}

	// res := &http.Response{}
	// res.Body = strings.NewReader(jout.Body)
	// TODO: shouldn't we pass back the full response object or something so we can set some things on it here?
	// For instance, user could set response content type or what have you.
	//io.Copy(cfg.Stdout, strings.NewReader(jout.Body))

	if rw, ok := w.(http.ResponseWriter); ok {
		b, err = json.Marshal(jout.Body)
		if err != nil {
			return fmt.Errorf("error unmarshalling JSONOutput.Body: %v", err)
		}
		rw.WriteHeader(jout.StatusCode)
		rw.Write(b) // TODO timeout
	} else {
		// logs can just copy the full thing in there, headers and all.
		b, err = json.Marshal(jout)
		if err != nil {
			return fmt.Errorf("error unmarshalling JSONOutput: %v", err)
		}

		w.Write(b) // TODO timeout

	}
	return nil

}
