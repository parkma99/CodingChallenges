package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const MAX_TIMEOUT_MS = 1000
const MAX_PARALLEL_NUM = 100
const MAX_PORT = 65536

var timeout int
var parallel_num int

type hostList []string

func (hl *hostList) String() string {
	return strings.Join(*hl, ",")
}

func (hl *hostList) Set(value string) error {
	hosts := strings.Split(strings.TrimSpace(value), ",")
	new_hl := make([]string, len(hosts))
	for i, host := range hosts {
		new_hl[i] = strings.TrimSpace(host)
	}
	*hl = new_hl
	return nil
}

func cmd() (hostList, int, error) {
	var hosts hostList
	var port int
	flag.Var(&hosts, "host", "host to connect")
	flag.IntVar(&port, "port", 0, "port to connect")
	flag.IntVar(&timeout, "timeout", MAX_TIMEOUT_MS, "max timeout microsecond")
	flag.IntVar(&parallel_num, "parallel", MAX_PARALLEL_NUM, "max parallel number")
	flag.Parse()
	if len(hosts) == 0 {
		return []string{}, port, errors.New("parse flag error")
	}
	return hosts, port, nil
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
	hosts, port, err := cmd()
	log.Println(hosts, port)
	if err != nil {
		log.Fatalf(err.Error())
	}
	if port != 0 {
		for _, host := range hosts {
			fmt.Fprintf(os.Stdout, "Scanning host: %s port: %d\n", []any{host, port}...)
			tryCreateTCPConnect(host, port)
		}

	} else {
		for _, host := range hosts {
			parallel_scan(host)
		}

	}
}

func parallel_scan(host string) {
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
