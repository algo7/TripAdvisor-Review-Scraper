package tripadvisor

import (
	"context"
	"fmt"
	"net"
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
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	return &http.Client{
		Transport: httpTransport,
	}, nil
}
