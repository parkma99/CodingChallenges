package main

import "strings"

type ccScanner interface {
	new(host string, timeout int, ports []uint16, parallel int) error
	checker(port uint16) bool
	scan() ([]uint16, error)
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
