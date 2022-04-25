package netutil

import (
	"fmt"
	"net"
)

func RandomAddr() (string, error) {
	const addr = "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("could not listen to %s: %w", addr, err)
	}
	defer l.Close()
	return l.Addr().String(), nil
}
