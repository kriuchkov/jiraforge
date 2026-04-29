package atlassian

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-faster/errors"
)

func TestNewHTTPClientAppliesSafeTimeouts(t *testing.T) {
	client, err := newHTTPClient("")
	if err != nil {
		t.Fatal(errors.Wrap(err, "newHTTPClient returned error"))
	}
	if client.Timeout != defaultHTTPTimeout {
		t.Fatal(errors.Errorf("unexpected client timeout: got %s want %s", client.Timeout, defaultHTTPTimeout))
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal(errors.Errorf("unexpected transport type: %T", client.Transport))
	}
	if transport.TLSHandshakeTimeout != defaultTLSHandshakeTimeout {
		t.Fatal(errors.Errorf("unexpected TLS handshake timeout: got %s want %s", transport.TLSHandshakeTimeout, defaultTLSHandshakeTimeout))
	}
	if transport.ResponseHeaderTimeout != defaultResponseHeaderTimeout {
		t.Fatal(errors.Errorf("unexpected response header timeout: got %s want %s", transport.ResponseHeaderTimeout, defaultResponseHeaderTimeout))
	}
	if transport.IdleConnTimeout != defaultIdleConnTimeout {
		t.Fatal(errors.Errorf("unexpected idle timeout: got %s want %s", transport.IdleConnTimeout, defaultIdleConnTimeout))
	}
	if transport.ExpectContinueTimeout != defaultExpectContinueTimeout {
		t.Fatal(errors.Errorf("unexpected expect-continue timeout: got %s want %s", transport.ExpectContinueTimeout, defaultExpectContinueTimeout))
	}
	if transport.MaxIdleConns != defaultMaxIdleConns {
		t.Fatal(errors.Errorf("unexpected max idle conns: got %d want %d", transport.MaxIdleConns, defaultMaxIdleConns))
	}
	if transport.MaxIdleConnsPerHost != defaultMaxIdleConnsPerHost {
		t.Fatal(errors.Errorf("unexpected max idle conns per host: got %d want %d", transport.MaxIdleConnsPerHost, defaultMaxIdleConnsPerHost))
	}
	if transport.ForceAttemptHTTP2 != true {
		t.Fatal(errors.New("expected HTTP/2 to be enabled"))
	}
	if transport.TLSClientConfig != nil {
		t.Fatal(errors.New("expected TLSClientConfig to be nil by default"))
	}

	dialer := transport.DialContext
	if dialer == nil {
		t.Fatal(errors.New("expected DialContext to be configured"))
	}
	if defaultDialTimeout <= 0 || defaultKeepAlive <= 0 || client.Timeout <= 0*time.Second {
		t.Fatal(errors.New("expected positive timeout defaults"))
	}
}

func TestNewHTTPClientUsesExplicitProxy(t *testing.T) {
	client, err := newHTTPClient("http://proxy.internal:8080")
	if err != nil {
		t.Fatal(errors.Wrap(err, "newHTTPClient returned error"))
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal(errors.Errorf("unexpected transport type: %T", client.Transport))
	}
	if transport.Proxy == nil {
		t.Fatal(errors.New("expected proxy function to be configured"))
	}

	request, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net", nil)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to build request"))
	}
	proxyURL, err := transport.Proxy(request)
	if err != nil {
		t.Fatal(errors.Wrap(err, "proxy resolution failed"))
	}
	if proxyURL == nil {
		t.Fatal(errors.New("expected proxy URL, got nil"))
	}
	if got := proxyURL.String(); got != "http://proxy.internal:8080" {
		t.Fatal(errors.Errorf("unexpected proxy URL: got %q", got))
	}
	if transport.TLSClientConfig != nil {
		t.Fatal(errors.New("expected explicit proxy to avoid insecure TLS overrides"))
	}
}

func TestNewHTTPClientRejectsInvalidProxyURL(t *testing.T) {
	if _, err := newHTTPClient("://bad proxy"); err == nil {
		t.Fatal(errors.New("expected invalid proxy URL to return an error"))
	}
}
