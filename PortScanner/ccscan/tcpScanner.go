package main

import (
	"fmt"
	"net"
)

type tcpScanner struct {
}

func (s *tcpScanner) checker(config *ccScanConfig, port uint16) bool {
	address := fmt.Sprintf("%s:%d", config.host.String(), port)
	conn, err := net.DialTimeout("tcp", address, config.timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
