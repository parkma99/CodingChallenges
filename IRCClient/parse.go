package main

import (
	"strings"
)

type IRCMessage struct {
	Raw                    string
	Nick, Ident, Host, Src string
	Cmd                    string
	Args                   []string
}

func parseIRCMessage(s string) *IRCMessage {
	msg := &IRCMessage{Raw: s}
	if s == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	if s[0] == ':' {
		if idx := strings.Index(s, " "); idx != -1 {
			msg.Src, s = s[1:idx], s[idx+1:]
		} else {
			return nil
		}

		msg.Host = msg.Src
		if n, i, h, ok := parseUserHost(msg.Src); ok {
			msg.Nick = n
			msg.Ident = i
			msg.Host = h
		}
	}
	args := strings.SplitN(s, " :", 2)
	if len(args) > 1 {
		args = append(strings.Fields(args[0]), args[1])
	} else {
		args = strings.Fields(args[0])
	}
	msg.Cmd = strings.ToUpper(args[0])
	if len(args) > 1 {
		msg.Args = args[1:]
	}
	return msg
}

func parseUserHost(uh string) (nick, ident, host string, ok bool) {
	uh = strings.TrimSpace(uh)
	nidx, uidx := strings.Index(uh, "!"), strings.Index(uh, "@")
	if uidx == -1 || nidx == -1 {
		return "", "", "", false
	}
	return uh[:nidx], uh[nidx+1 : uidx], uh[uidx+1:], true
}
