package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type connect struct {
	Conn          net.Conn
	Server        string
	Port          string
	Reader        *bufio.Reader
	Writer        *bufio.Writer
	InputCmdChan  chan string
	ServerMsgChan chan string
}

func NewConnect() (c *connect, err error) {
	c = &connect{}
	return c, nil
}
func closeConnection(conn *connect) {
	if conn.Conn != nil {
		conn.Writer.Flush()
		conn.Conn.Close()
		conn.Conn = nil
		close(conn.InputCmdChan)
		close(conn.ServerMsgChan)
		log.Println("exit client graceful")
	}
}

func (conn *connect) readMessages() {
	defer closeConnection(conn)

	for {
		message, err := conn.Reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			return
		}
		conn.ServerMsgChan <- message
	}
}

func (conn *connect) writeMessages() {
	defer closeConnection(conn)

	for {
		select {
		case msg := <-conn.InputCmdChan:
			if conn.Conn == nil {
				return
			}
			log.Println("ready to write", msg)
			_, err := conn.Writer.WriteString(msg + "\r\n")
			if err != nil {
				return
			}
			err = conn.Writer.Flush()
			if err != nil {
				return
			}
		case text := <-conn.ServerMsgChan:
			msg := parseIRCMessage(text)
			if msg == nil {
				return
			}
			log.Println(text)
			log.Println(msg.Nick, msg.Ident, msg.Src, msg.Host)
			log.Println(msg.Cmd, len(msg.Args), msg.Args)
			if msg.Cmd == "PING" {
				resp := fmt.Sprintf("PONG :%s\r\n", msg.Args[0])
				conn.Writer.WriteString(resp)
				conn.Writer.Flush()
			}
		}
	}
}
