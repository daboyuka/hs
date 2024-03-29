package httpcmd

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

type flakyTransport struct {
	FailProb float64
	Next     http.RoundTripper
}

func (f flakyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if rand.Float64() < f.FailProb {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Status:     "Service Unavailable",
			Body:       io.NopCloser(strings.NewReader("wow everything exploded")),
		}, nil
	}
	return f.Next.RoundTrip(request)
}
