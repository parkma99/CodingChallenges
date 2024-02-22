package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
)

func cmd() (string, int, error) {
	var host string
	var port int
	flag.StringVar(&host, "host", "", "host to connect")
	flag.IntVar(&port, "port", 0, "port to connect")
	flag.Parse()
	if host == "" {
		return "", port, errors.New("parse flag error")
	}
	return host, port, nil
}

func tryCreateTCPConnect(host string, port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
func main() {
	host, port, err := cmd()
	if err != nil {
		log.Fatalf(err.Error())
	}
	if port == 0 {
		fmt.Printf("Scanning host: %s \n", host)
		for i := range 65536 {
			if tryCreateTCPConnect(host, i+1) {
				fmt.Printf("Port: %d is open\n", i+1)
			}
		}
	} else {
		fmt.Printf("Scanning host: %s port: %d\n", host, port)
		if tryCreateTCPConnect(host, port) {
			fmt.Printf("Port: %d is open\n", port)
		}
	}

}
