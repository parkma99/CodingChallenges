package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
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
		err := scanner.new(host, timeout, ports, parallel)
		if err != nil {
			log.Fatalln(err.Error())
		}
		opened, err := scanner.scan()
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
