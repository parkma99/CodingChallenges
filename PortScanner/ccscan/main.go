package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
)

const MAX_TIMEOUT_MS = 50000
const MAX_PARALLEL = 100
const MAX_PORT = 65535
const DEFAULT_TYPE = "TCP"

func main() {
	var hosts hostList
	var port uint
	var scanType string
	var timeout int
	var parallel int
	flag.Var(&hosts, "host", "host to connect")
	flag.UintVar(&port, "port", 0, "port to connect")
	flag.IntVar(&timeout, "timeout", MAX_TIMEOUT_MS, "max timeout microsecond")
	flag.IntVar(&parallel, "parallel", MAX_PARALLEL, "max parallel number")
	flag.StringVar(&scanType, "type", DEFAULT_TYPE, "scanner type, syn need root")
	flag.Parse()
	if len(hosts) == 0 {
		return
	}

	ports := make([]uint16, 0)
	if port != 0 {
		ports = []uint16{uint16(port)}
	} else {
		for i := range MAX_PORT {
			ports = append(ports, uint16(i+1))
		}
	}
	log.Println(hosts)
	for _, host := range hosts {
		fmt.Println("scanning ", host)
		scanner := createScanner(scanType)
		if scanner == nil {
			return
		}
		err := scanner.new(host, timeout)
		if err != nil {
			log.Fatalln(err.Error())
		}
		opened := startScanner(scanner, ports, parallel)
		if err != nil {
			continue
		}
		sort.Slice(opened, func(i, j int) bool {
			return opened[i] < opened[j]
		})
		for _, port := range opened {
			fmt.Println(port, "is opened")
		}
	}
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
func startScanner(s ccScanner, checkPorts []uint16, parallel int) (ports []uint16) {
	parallelNum := len(checkPorts)
	if parallel > 0 && parallel < parallelNum {
		parallelNum = parallel
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
		for _, p := range checkPorts {
			inputs <- p
		}
	}()

	for i := 0; i < len(checkPorts); i++ {
		port := <-results
		if port != 0 {
			ports = append(ports, port)
		}
	}
	close(inputs)
	close(results)
	return
}
