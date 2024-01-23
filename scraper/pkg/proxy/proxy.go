package proxy

import (
	"fmt"
	"net/http"

	"golang.org/x/net/proxy"
)

// GetHTTPClientWithProxy returns an HTTP client that uses the proxy server
func GetHTTPClientWithProxy(proxyHost string) (*http.Client, error) {
	// Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyHost, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("error creating dialer: %w", err)
	}

	// Configure the HTTP client to use the dialer
	httpTransport := &http.Transport{
		// Set the Dial function to the dialer.Dial method
		Dial: dialer.Dial,
	}

	return &http.Client{
		Transport: httpTransport,
	}, nil
}
