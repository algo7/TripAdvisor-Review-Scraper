package tripadvisor

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// GetHTTPClientWithProxy returns an HTTP client that uses the proxy server
func GetHTTPClientWithProxy(proxyHost string) (*http.Client, error) {

	// Check if the proxy server is operational
	if !CheckProxyConnection(proxyHost, 5*time.Second) {
		return nil, fmt.Errorf("proxy server is not operational")
	}
	log.Println("Proxy server is operational")

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
		Timeout:   10 * time.Second,
	}, nil
}

// CheckProxyConnection attempts to establish a TCP connection to the proxy server
// to determine if it is operational.
func CheckProxyConnection(proxyHost string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", proxyHost, timeout)
	if err != nil {
		fmt.Printf("Error connecting to proxy: %v\n", err)
		return false
	}
	defer conn.Close()
	return true
}
