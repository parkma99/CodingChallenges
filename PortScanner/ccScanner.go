package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type ccScanner interface {
	checker(config *ccScanConfig, port uint16) bool
}
type ccScanConfig struct {
	host       net.IP
	timeout    time.Duration
	checkPorts []uint16
	parallel   int
}

func newConfig(host string, timeout int, ports []uint16, parallel int) (*ccScanConfig, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("parese %s to ip error", host)
	}

	return &ccScanConfig{
		host:       ip,
		timeout:    time.Duration(timeout) * time.Microsecond,
		checkPorts: ports,
		parallel:   parallel,
	}, nil
}

func createScanner(scanType string) ccScanner {
	switch strings.ToLower(scanType) {
	case "tcp":
		return &tcpScanner{}
	case "syn":
		return &synScanner{}
	default:
		return nil
	}
}

func scan(s ccScanner, config *ccScanConfig) (ports []uint16, err error) {
	parallelNum := len(config.checkPorts)
	if config.parallel > 0 && config.parallel < parallelNum {
		parallelNum = config.parallel
	}
	inputs := make(chan uint16, 100)
	results := make(chan uint16)

	for i := 0; i < parallelNum; i++ {
		go func(ports <-chan uint16, result chan<- uint16) {
			for port := range ports {
				opened := s.checker(config, port)
				if opened {
					result <- port
				} else {
					result <- 0
				}
			}
		}(inputs, results)
	}
	go func() {
		for _, p := range config.checkPorts {
			inputs <- p
		}
	}()

	for i := 0; i < len(config.checkPorts); i++ {
		port := <-results
		if port != 0 {
			ports = append(ports, port)
		}
	}
	close(inputs)
	close(results)
	return
}
