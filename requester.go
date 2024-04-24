package requester

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	url        string
	method     string
	err        error
	body       io.Reader
	header     http.Header
	httpClient *http.Client
	response   *http.Response

	handlers []handler
}

func New(method, url string) *Request {
	return &Request{
		url:        url,
		method:     method,
		header:     make(http.Header),
		httpClient: &http.Client{Timeout: 1 * time.Minute},
	}
}

func POST(url string) *Request {
	return &Request{
		url:        url,
		method:     http.MethodPost,
		header:     make(http.Header),
		httpClient: &http.Client{Timeout: 1 * time.Minute},
	}
}

func GET(url string) *Request {
	return &Request{
		url:        url,
		method:     http.MethodGet,
		header:     make(http.Header),
		httpClient: &http.Client{Timeout: 1 * time.Minute},
	}
}

func (r *Request) Proxy(url *url.URL) *Request {
	if r.err != nil {
		return r
	}
	r.httpClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(url),
	}
	return r
}

func (r *Request) Timeout(dur time.Duration) *Request {
	if r.err != nil {
		return r
	}
	r.httpClient.Timeout = dur
	return r
}

func (r *Request) Header(header http.Header) *Request {
	if r.err != nil {
		return r
	}
	r.header = header
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.err != nil {
		return r
	}
	r.header.Set(key, value)
	return r
}

func (r *Request) ContentType(contentType string) *Request {
	if r.err != nil {
		return r
	}
	r.header.Set("Content-Type", contentType)
	return r
}

func (r *Request) HTTPClient(client *http.Client) *Request {
	if r.err != nil {
		return r
	}
	r.httpClient = client
	return r
}

func (r *Request) Body(body string) *Request {
	if r.err != nil {
		return r
	}
	return &Request{body: strings.NewReader(body)}
}

func (r *Request) BodyBytes(body []byte) *Request {
	if r.err != nil {
		return r
	}
	return &Request{body: bytes.NewBuffer(body)}
}

func (r *Request) BodyMarshal(body any) *Request {
	if r.err != nil {
		return r
	}
	b, err := json.Marshal(body)
	if err != nil {
		r.err = &Error{Op: "BodyMarshal", Err: err.Error()}
	}
	r.body = bytes.NewBuffer(b)
	return r
}

func (r *Request) Do() error {
	if r.err != nil {
		return r.err
	}
	req, err := r.newRequest()
	if err != nil {
		return err
	}
	r.response, err = r.do(req)
	if err != nil {
		return err
	}
	r.chainHandlers()
	return r.err
}

func (r *Request) chainHandlers() {
	for _, h := range r.handlers {
		if h == nil {
			continue
		}
		h(r)
	}
}

func (r *Request) DoWithContext(ctx context.Context) error {
	req, err := r.newRequestWithContext(ctx)
	if err != nil {
		return err
	}
	r.response, err = r.do(req)
	if err != nil {
		return err
	}
	r.chainHandlers()
	return nil
}

func (r *Request) closeBody() {
	if r.response.Body != nil {
		if err := r.response.Body.Close(); err != nil {
			r.err = &Error{Op: "closeBody", Err: err.Error()}
		}
	}
}

func (r *Request) do(req *http.Request) (*http.Response, error) {
	return r.httpClient.Do(req)
}

func (r *Request) newRequest() (*http.Request, error) {
	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		return nil, err
	}
	req.Header = r.header
	return req, nil
}

func (r *Request) newRequestWithContext(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, r.body)
	if err != nil {
		return nil, err
	}
	req.Header = r.header
	return req, nil
}
