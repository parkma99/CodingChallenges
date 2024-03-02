package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	conn, err := NewConnect()
	if err != nil {
		log.Fatalln(err)
	}
	p := tea.NewProgram(initialModel(conn))
	log.Println(conn)
	go func() {
		for resp := range conn.ServerMsgChan {
			message := parseIRCMessage(resp)
			if message != nil {
				if message.Cmd == "PING" {
					conn.InputCmdChan <- fmt.Sprintf("PONG :%s\r\n", message.Args[0])
				} else if message.Cmd != "372" {
					// log.Println("send to tui", message.Raw)
					p.Send(*message)
				}
			}
		}
		p.Send(tea.KeyMsg(tea.Key{Type: tea.KeyEsc}))
	}()
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
