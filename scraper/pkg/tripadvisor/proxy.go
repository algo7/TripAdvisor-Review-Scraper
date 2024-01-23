package tripadvisor

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// GetHTTPClientWithProxy returns an HTTP client that uses the proxy server
func GetHTTPClientWithProxy(proxyHost string) (*http.Client, error) {

	// Parse the proxy URL into a URL struct
	proxyURL, err := url.Parse(proxyHost)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy URL: %w", err)
	}

	// Check if the proxy server is operational
	if !CheckProxyConnection(proxyURL.Host, 5*time.Second) {
		return nil, fmt.Errorf("proxy server is not operational")
	}
	log.Println("Proxy server is operational")

	// Create a new HTTP transport with the proxy URL
	httpTransport := &http.Transport{
		Proxy:             http.ProxyURL(proxyURL),
		ForceAttemptHTTP2: true,
		// If set to true, new connections will be established to the proxy server every time.
		DisableKeepAlives: false,
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
