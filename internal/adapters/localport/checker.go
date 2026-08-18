package localport

import (
	"net"
	"strconv"
)

type Checker struct{}

func (Checker) Available(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}
