package main

import (
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type synScanner struct {
	host       net.IP
	timeout    time.Duration
	checkPorts []uint16
	parallel   int
}

func (s *synScanner) new(host string, timeout int, ports []uint16, parallel int) (err error) {
	s.host = net.ParseIP(host)
	s.checkPorts = ports
	s.parallel = parallel
	s.timeout = time.Duration(timeout) * time.Microsecond
	return nil
}
func (s *synScanner) scan() (ports []uint16, err error) {
	parallelNum := len(s.checkPorts)
	if s.parallel > 0 && s.parallel < parallelNum {
		parallelNum = s.parallel
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
		for _, p := range s.checkPorts {
			inputs <- p
		}
	}()

	for i := 0; i < len(s.checkPorts); i++ {
		port := <-results
		if port != 0 {
			ports = append(ports, port)
		}
	}
	close(inputs)
	close(results)
	return
}

func localIPPort(dstip net.IP) (net.IP, int) {
	serverAddr, err := net.ResolveUDPAddr("udp", dstip.String()+":12345")
	if err != nil {
		log.Fatal(err)
	}
	if con, err := net.DialUDP("udp", nil, serverAddr); err == nil {
		if udpaddr, ok := con.LocalAddr().(*net.UDPAddr); ok {
			return udpaddr.IP, udpaddr.Port
		}
	}
	log.Fatal("could not get local ip: " + err.Error())
	return nil, -1
}

func (s *synScanner) checker(port uint16) bool {

	dstip := s.host
	dstport := layers.TCPPort(port)
	srcip, sport := localIPPort(dstip)
	srcport := layers.TCPPort(sport)

	ip := &layers.IPv4{
		SrcIP:    srcip,
		DstIP:    dstip,
		Protocol: layers.IPProtocolTCP,
	}

	tcp := &layers.TCP{
		SrcPort: srcport,
		DstPort: dstport,
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		log.Println(err)
		return false
	}

	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Println(err)
		return false
	}
	defer conn.Close()
	if _, err := conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstip}); err != nil {
		log.Println(err)
		return false
	}

	// Set deadline so we don't wait forever.
	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Println(err)
		return false
	}

	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println("error reading packet: ", err)
			return false
		} else if addr.String() == dstip.String() {
			// Decode a packet
			packet := gopacket.NewPacket(buffer[:n], layers.LayerTypeTCP, gopacket.Default)
			// Get the TCP layer from this packet
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)

				if tcp.DstPort == srcport {
					return tcp.SYN && tcp.ACK
				}
			}
		}
	}
}
