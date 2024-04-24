package requester

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type handler func(r *Request)

func (r *Request) ExpectCode(code int) *Request {
	if r.err != nil {
		return r
	}
	r.handlers = append(r.handlers, expectCodeHandler(code))
	return r
}

func (r *Request) ToCookies(cookie *[]*http.Cookie) *Request {
	if r.err != nil {
		return r
	}
	r.handlers = append(r.handlers, cookiesHandler(cookie))
	return r
}

func (r *Request) Decode(v any) *Request {
	if r.err != nil {
		return r
	}
	r.handlers = append(r.handlers, decodeToAnyHandler(v))
	return r
}

func (r *Request) ToString(v *string) *Request {
	if r.err != nil {
		return r
	}
	r.handlers = append(r.handlers, decodeToStringHandler(v))
	return r
}

func (r *Request) ToBytes(v *[]byte) *Request {
	if r.err != nil {
		return r
	}
	r.handlers = append(r.handlers, decodeToBytesHandler(v))
	return r
}

func cookiesHandler(cookies *[]*http.Cookie) handler {
	return func(r *Request) {
		if r.response == nil {
			r.err = &Error{Op: "ToCookies", Err: "response is nil"}
			return
		}
		*cookies = r.response.Cookies()
	}
}

func decodeToAnyHandler(v any) handler {
	return func(r *Request) {
		if r.response == nil {
			r.err = &Error{Op: "Decode", Err: "response is nil"}
			return
		}

		defer r.closeBody()
		if err := json.NewDecoder(r.response.Body).Decode(v); err != nil {
			r.err = &Error{Op: "Decode", Err: err.Error()}
		}
	}
}

func decodeToStringHandler(v *string) handler {
	return func(r *Request) {
		if r.response == nil {
			r.err = &Error{Op: "ToString", Err: "response is nil"}
			return
		}

		defer r.closeBody()
		bodyInBytes, err := io.ReadAll(r.response.Body)
		if err != nil {
			r.err = &Error{Op: "ToJSON", Err: err.Error()}
			return
		}
		*v = string(bodyInBytes)
	}
}

func decodeToBytesHandler(v *[]byte) handler {
	return func(r *Request) {
		if r.response == nil {
			r.err = &Error{Op: "ToBytes", Err: "response is nil"}
			return
		}

		defer r.closeBody()
		bodyInBytes, err := io.ReadAll(r.response.Body)
		if err != nil {
			r.err = &Error{Op: "ToBytes", Err: err.Error()}
			return
		}
		*v = bodyInBytes
	}
}

func expectCodeHandler(code int) handler {
	return func(r *Request) {
		if r.response == nil {
			r.err = &Error{Op: "ExpectCode", Err: "response is nil"}
			return
		}
		if r.response.StatusCode != code {
			r.err = &Error{Op: "ExpectCode", Err: fmt.Sprintf("expect code %d, got %d", code, r.response.StatusCode)}
		}
	}
}
