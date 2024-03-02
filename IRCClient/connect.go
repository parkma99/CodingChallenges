package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const (
	DEFAULT_SERVER = "sakura.jp.as.dal.net"
	DEFAULT_PORT   = 6667
)

type connect struct {
	Conn          net.Conn
	Server        string
	Port          string
	Channel       string
	Reader        *bufio.Reader
	Writer        *bufio.Writer
	InputCmdChan  chan string
	ServerMsgChan chan string
}

func NewConnect() (c *connect, err error) {
	conn := &connect{}

	conn.Conn, _ = net.Dial("tcp", fmt.Sprintf("%s:%d", DEFAULT_SERVER, DEFAULT_PORT))
	conn.Reader = bufio.NewReader(conn.Conn)
	conn.Writer = bufio.NewWriter(conn.Conn)
	conn.InputCmdChan = make(chan string, 1)
	conn.ServerMsgChan = make(chan string, 1)
	nick := "CCClient"
	user := "parkma99"
	conn.Writer.WriteString(fmt.Sprintf("NICK %s\r\n", nick))
	conn.Writer.WriteString(fmt.Sprintf("USER %s 0 * :%s\r\n", user, user))
	conn.Writer.Flush()
	go func() {
		conn.readMessages()
	}()

	go func() {
		conn.writeMessages()
	}()
	return conn, nil
}
func closeConnection(conn *connect) {
	if conn.Conn != nil {
		conn.Writer.Flush()
		conn.Conn.Close()
		conn.Conn = nil
		close(conn.InputCmdChan)
		close(conn.ServerMsgChan)
	}
}

func (conn *connect) readMessages() {
	defer closeConnection(conn)

	for {
		if conn.Conn == nil {
			return
		}
		message, err := conn.Reader.ReadString('\n')
		// log.Println(message)
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
		if conn.Conn == nil {
			return
		}
		select {
		case msg := <-conn.InputCmdChan:
			_, err := conn.Writer.WriteString(msg)
			if err != nil {
				return
			}
			err = conn.Writer.Flush()
			if err != nil {
				return
			}
		default:
		}

	}
}

func (c *connect) parse2Cmd(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	args := strings.Split(input, " ")
	if len(args) == 0 {
		return ""
	}
	switch strings.ToLower(args[0]) {
	case "/join":
		if len(args) < 2 {
			return ""
		}
		c.Channel = args[1]
		return fmt.Sprintf("JOIN %s\r\n", args[1])
	case "/part":
		if len(args) < 2 {
			return fmt.Sprintf("PART %s\r\n", c.Channel)
		}
		return fmt.Sprintf("PART %s\r\n", args[1])
	case "/nick":
		if len(args) < 2 {
			return ""
		}
		return fmt.Sprintf("NICK %s\r\n", args[1])
	case "/quit":
		if len(args) < 2 {
			return "QUIT\r\n"
		}
		return fmt.Sprintf("QUIT :%s\r\n", args[1])
	default:
		return fmt.Sprintf("PRIVMSG #tests :%s\r\n", input)
	}
}
