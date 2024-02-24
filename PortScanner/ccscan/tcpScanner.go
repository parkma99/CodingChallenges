package main

import (
	"fmt"
	"net"
	"time"
)

type tcpScanner struct {
	config scannerConfig
}

func (s *tcpScanner) new(host string, timeout int) (err error) {
	s.config = scannerConfig{host: host, timeout: time.Duration(timeout) * time.Microsecond}
	return nil
}

func (s *tcpScanner) checker(port uint16) bool {
	address := fmt.Sprintf("%s:%d", s.config.host, port)
	// log.Println(address)
	conn, err := net.DialTimeout("tcp", address, s.config.timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
