package httpclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-pkgz/repeater"
	"github.com/go-pkgz/repeater/strategy"
	"github.com/pkg/errors"
)

const (
	httpTimeout         = 15 * time.Second
	dialTimeout         = 5 * time.Second
	tlsHandshakeTimeout = 5 * time.Second
	repeaterFactor      = 1.5
	repeatTimes         = 5
)

// Do makes retryable HTTP requests with predefined timeouts
func Do(ctx context.Context, req *http.Request) (io.ReadCloser, error) {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: dialTimeout,
		}).Dial,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
	}
	client := &http.Client{
		Timeout:   httpTimeout,
		Transport: transport,
	}

	var res io.ReadCloser
	f := func() error {
		resp, ferr := client.Do(req)
		if ferr != nil {
			return errors.Wrap(ferr, "Error making HTTP request")
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			defer resp.Body.Close()
			return errors.New(fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		res = resp.Body
		return nil
	}

	r := repeater.New(&strategy.Backoff{
		Repeats: repeatTimes,
		Factor:  repeaterFactor,
		Jitter:  true,
	})
	if err := r.Do(ctx, f); err != nil {
		return nil, errors.Wrap(err, "repeater tried hard, but no luck")
	}

	return res, nil
}
