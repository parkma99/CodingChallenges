package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
)

type synScanner struct {
	config scannerConfig
	iface  *net.Interface
	dst    net.IP
	gw     net.IP
	src    net.IP
	opts   gopacket.SerializeOptions
	buf    gopacket.SerializeBuffer
}

func (s *synScanner) new(host string, timeout int) (err error) {
	s.config = scannerConfig{host: host, timeout: time.Duration(timeout) * time.Microsecond}
	router, err := routing.New()
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("non-ip target: %q", host)
	}
	s.dst = ip
	s.opts = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	s.buf = gopacket.NewSerializeBuffer()
	iface, gw, src, err := router.Route(ip)
	if err != nil {
		return err
	}
	s.gw, s.src, s.iface = gw, src, iface

	return
}
func parseMac(macaddr string) net.HardwareAddr {
	parsedMac, _ := net.ParseMAC(macaddr)
	return parsedMac
}
func (s *synScanner) checker(port uint16) bool {
	hwaddr := parseMac("00:00:00:00:00:00")
	handle, err := pcap.OpenLive(s.iface.Name, MAX_PORT, true, pcap.BlockForever)
	if err != nil {
		return false
	}
	defer handle.Close()
	if err != nil {
		return false
	}
	if err != nil {
		return false
	}
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    s.src,
		DstIP:    s.dst,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	srcPort, err := getFreePort()
	if err != nil {
		log.Fatalln(err.Error())
	}
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(port),
		SYN:     true,
	}
	err = tcp.SetNetworkLayerForChecksum(&ip4)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// Create the flow we expect returning packets to have, so we can check
	// against it and discard useless packets.
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, s.dst, s.src)
	start := time.Now()
	if err := gopacket.SerializeLayers(s.buf, s.opts, &eth, &ip4, &tcp); err != nil {
		return false
	}
	if err := handle.WritePacketData(s.buf.Bytes()); err != nil {
		return false
	}
	// Read in the next packet.
	data, _, err := handle.ReadPacketData()
	if err == pcap.NextErrorTimeoutExpired {
	} else if err != nil {
		log.Printf("error reading packet: %v", err)
	}

	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)

	if net := packet.NetworkLayer(); net == nil {
		// log.Printf("packet has no network layer")
	} else if net.NetworkFlow() != ipFlow {
		// log.Printf("packet does not match our ip src/dst")
	} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
		// log.Printf("packet has not tcp layer")
	} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
		// We panic here because this is guaranteed to never
		// happen.
		panic("tcp layer is not tcp layer :-/")
	} else if tcp.DstPort != layers.TCPPort(srcPort) {
		// log.Printf("dst port %v does not match", tcp.DstPort)
	} else if tcp.RST {
		// log.Printf("  port %v closed", tcp.SrcPort)
	} else if tcp.SYN && tcp.ACK {
		return true
	} else {
		return false
	}
	if time.Since(start) > s.config.timeout {
		return false
	}
	return false
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
