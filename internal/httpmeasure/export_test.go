package httpmeasure

import (
	"context"
	"crypto/x509"
	"net"
)

// ControlledTLSClient exposes the existing socket boundary only in test builds.
// TLS, HTTP, selection policy and orchestration remain production code.
func ControlledTLSClient(roots *x509.CertPool, dial func(context.Context, string, string) (net.Conn, error)) *Client {
	return &Client{Roots: roots, dial: dial}
}
