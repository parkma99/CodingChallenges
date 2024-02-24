package main

import (
	"fmt"
	"net"
	"time"
)

type tcpScanner struct {
	host       net.IP
	timeout    time.Duration
	checkPorts []uint16
	parallel   int
}

func (s *tcpScanner) new(host string, timeout int, ports []uint16, parallel int) (err error) {
	s.host = net.ParseIP(host)
	s.checkPorts = ports
	s.parallel = parallel
	s.timeout = time.Duration(timeout) * time.Microsecond
	return nil
}
func (s *tcpScanner) scan() (ports []uint16, err error) {
	parallelNum := len(s.checkPorts)
	if s.parallel > 0 && s.parallel < parallelNum {
		parallelNum = s.parallel
	}
	inputs := make(chan uint16, 100)
	results := make(chan uint16)

	for i := 0; i < parallelNum; i++ {
		go func(ports <-chan uint16, result chan<- uint16) {
			for port := range ports {
				opened := s.checker(port)
				if opened {
					result <- port
				} else {
					result <- 0
				}
			}
		}(inputs, results)
	}
	go func() {
		for _, p := range s.checkPorts {
			inputs <- p
		}
	}()

	for i := 0; i < len(s.checkPorts); i++ {
		port := <-results
		if port != 0 {
			ports = append(ports, port)
		}
	}
	close(inputs)
	close(results)
	return
}

func (s *tcpScanner) checker(port uint16) bool {
	address := fmt.Sprintf("%s:%d", s.host.String(), port)
	conn, err := net.DialTimeout("tcp", address, s.timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
