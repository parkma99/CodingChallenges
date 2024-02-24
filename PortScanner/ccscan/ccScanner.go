package main

import "time"

type ccScanner interface {
	new(host string, timeout int) error
	checker(port uint16) bool
}

type scannerConfig struct {
	host    string
	timeout time.Duration
}
