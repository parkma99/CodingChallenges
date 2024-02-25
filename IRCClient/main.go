package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	DEFAULT_SERVER = "sakura.jp.as.dal.net"
	DEFAULT_PORT   = 6667
)

func main() {
	conn, _ := NewConnect()
	var wg sync.WaitGroup

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

	wg.Add(2)
	go func() {
		defer wg.Done()
		conn.readMessages()
	}()

	go func() {
		defer wg.Done()
		conn.writeMessages()
	}()
	for {
		fmt.Print("Enter command: ")
		userInput := ""
		_, err := fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		userInput = strings.TrimSpace(userInput)
		if userInput == "join" {
			cmd := fmt.Sprintln("JOIN #tests")
			log.Println(cmd)
			conn.InputCmdChan <- cmd

		} else if userInput == "part" {
			cmd := fmt.Sprintln("PART #tests")
			log.Println(cmd)
			conn.InputCmdChan <- cmd
		} else if userInput == "nick" {
			conn.InputCmdChan <- "NICK CCTEST"
		} else if userInput == "quit" {
			conn.InputCmdChan <- "QUIT :Gone to have lunch"
		} else {
			userInput = strings.TrimSpace(userInput)
			cmd := fmt.Sprintf("PRIVMSG #tests :%s", userInput)
			conn.InputCmdChan <- cmd
		}
	}
}
