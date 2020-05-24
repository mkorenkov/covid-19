package httpclient

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-pkgz/repeater"
	"github.com/go-pkgz/repeater/strategy"
	"github.com/pkg/errors"
)

const (
	httpTimeout         = 10 * time.Second
	dialTimeout         = 5 * time.Second
	tlsHandshakeTimeout = 5 * time.Second
	repeaterFactor      = 1.5
	repeatTimes         = 5
)

var client *http.Client

func init() {
	var httpTransport *http.Transport
	if tr, ok := http.DefaultTransport.(*http.Transport); ok {
		// NOTE: non-racy copy of http transport
		// https://groups.google.com/forum/#!topic/golang-nuts/JmpHoAd76aU
		// https://github.com/golang/go/issues/26013
		// https://go-review.googlesource.com/c/go/+/174597/5/src/net/http/transport.go#295
		httpTransport = tr.Clone()
	} else {
		panic("http.DefaultTransport is not (*http.Transport). net/http changed in stdlib.")
	}
	// see default values here: https://golang.org/pkg/net/http/#RoundTripper
	httpTransport.DialContext = (&net.Dialer{
		Timeout:   httpTimeout,
		KeepAlive: 2 * httpTimeout,
		DualStack: true,
	}).DialContext
	httpTransport.TLSHandshakeTimeout = tlsHandshakeTimeout
	client = &http.Client{
		Timeout:   httpTimeout,
		Transport: httpTransport,
	}
}

// HTTPClient common interface for HTTP clients.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Retryable returns HTTP client with sane defaults, that can retry HTTP calls.
// Use `MakeRetryable` if you want to use your own HTTP Client.
func Retryable() HTTPClient {
	return &retryableClient{Client: client}
}

// MakeRetryable accepts HTTP client and makes it retryable.
func MakeRetryable(c HTTPClient) HTTPClient {
	return &retryableClient{Client: c}
}

type retryableClient struct {
	Client HTTPClient
}

// Do makes retryable HTTP requests with predefined timeouts
func (r *retryableClient) Do(req *http.Request) (res *http.Response, err error) {
	f := func() error {
		resp, ferr := client.Do(req)
		if ferr != nil {
			return errors.Wrap(ferr, "Error making HTTP request")
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			defer resp.Body.Close()
			return errors.New(fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		res = resp
		return nil
	}

	rp := repeater.New(&strategy.Backoff{
		Repeats: repeatTimes,
		Factor:  repeaterFactor,
		Jitter:  true,
	})
	if err := rp.Do(req.Context(), f); err != nil {
		return nil, errors.Wrap(err, "repeater tried hard, but no luck")
	}

	return res, nil
}
