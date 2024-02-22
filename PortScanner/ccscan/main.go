package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const MAX_TIMEOUT_MS = 1000
const MAX_PARALLEL_NUM = 100
const MAX_PORT = 65536

var timeout int
var parallel_num int

func cmd() (string, int, error) {
	var host string
	var port int
	flag.StringVar(&host, "host", "", "host to connect")
	flag.IntVar(&port, "port", 0, "port to connect")
	flag.IntVar(&timeout, "timeout", MAX_TIMEOUT_MS, "max timeout microsecond")
	flag.IntVar(&parallel_num, "parallel", MAX_PARALLEL_NUM, "max parallel number")
	flag.Parse()
	if host == "" {
		return "", port, errors.New("parse flag error")
	}
	return host, port, nil
}

func tryCreateTCPConnect(host string, port int) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(timeout)*time.Microsecond)
	if err != nil {
		return
	}
	conn.Close()
	fmt.Printf("Port: %d is open\n", port)
}
func main() {
	host, port, err := cmd()
	if err != nil {
		log.Fatalf(err.Error())
	}
	if port != 0 {
		fmt.Fprintf(os.Stdout, "Scanning host: %s port: %d\n", []any{host, port}...)
		tryCreateTCPConnect(host, port)
	} else {
		fmt.Printf("Scanning host: %s \n", host)
		low := 0
		high := 0
		step := MAX_PORT / parallel_num
		done := make(chan bool)
		for range parallel_num {
			low = high + 1
			if low > MAX_PORT {
				break
			}
			high = low + step
			if high >= MAX_PORT {
				high = MAX_PORT
			}
			go func(low, high int) {
				for i := low; i <= high; i++ {
					tryCreateTCPConnect(host, i)
				}
				done <- true
			}(low, high)
		}
		for range parallel_num {
			<-done
		}
	}

}
