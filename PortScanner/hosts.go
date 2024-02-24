package main

import (
	"net"
	"net/netip"
	"strings"
)

func cidr(cidr string) ([]string, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, err
	}
	var ips []string
	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		ips = append(ips, addr.String())
	}
	if len(ips) < 2 {
		return ips, nil
	}
	return ips[:len(ips)-1], nil
}

type hostList []string

func (hl *hostList) String() string {
	return strings.Join(*hl, ",")
}

func domain2IP(domain string) (ips []string, err error) {
	ip := net.ParseIP(domain)
	if ip != nil {
		return []string{domain}, nil
	}
	res, err := net.LookupIP(domain)
	if err != nil {
		return
	}
	for _, ip := range res {
		if ip4 := ip.To4(); ip4 != nil {
			ips = append(ips, ip4.String())
		}

	}
	return
}

func (hl *hostList) Set(value string) error {
	hosts := strings.Split(strings.TrimSpace(value), ",")
	new_hl := make([]string, 0)
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if strings.Contains(host, "/") {
			matches, err := cidr(host)
			if err != nil {
				continue
			}
			for _, addr := range matches {
				ips, err := domain2IP(addr)
				if err != nil {
					continue
				}
				new_hl = append(new_hl, ips...)
			}
		} else {
			ips, err := domain2IP(host)
			if err != nil {
				continue
			}
			new_hl = append(new_hl, ips...)
		}
	}
	*hl = new_hl
	return nil
}
