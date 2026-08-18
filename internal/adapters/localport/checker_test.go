package localport

import (
	"net"
	"testing"
)

func TestCheckerReportsBoundPortUnavailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on temporary port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	if (Checker{}).Available(port) {
		t.Fatalf("expected bound port %d to be unavailable", port)
	}
}
