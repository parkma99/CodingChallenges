package main

import (
	"testing"
)

func TestParseUserHost(t *testing.T) {
	tests := []struct {
		in, nick, ident, host string
		ok                    bool
	}{
		{"", "", "", "", false},
		{"   ", "", "", "", false},
		{"somestring", "", "", "", false},
		{" s p ", "", "", "", false},
		{"foo!bar", "", "", "", false},
		{"foo@baz.com", "", "", "", false},
		{"foo!bar@baz.com", "foo", "bar", "baz.com", true},
		{"  foo!bar@baz.com", "foo", "bar", "baz.com", true},
		{" foo!bar@baz.com  ", "foo", "bar", "baz.com", true},
	}

	for i, test := range tests {
		nick, ident, host, ok := parseUserHost(test.in)
		if test.nick != nick ||
			test.ident != ident ||
			test.host != host ||
			test.ok != ok {
			t.Errorf("%d: parseUserHost(%q) = %q, %q, %q, %t; want %q, %q, %q, %t",
				i, test.in, nick, ident, host, ok, test.nick, test.ident, test.host, test.ok)
		}
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		in  string
		msg *IRCMessage
		ok  bool
	}{
		{":*.freenode.net 353 CCIRC = #cc :@CCIRC", &IRCMessage{}, false},
		{":*.freenode.net NOTICE CCIRC :*** Ident lookup timed out, using ~guest instead.", &IRCMessage{}, false},
		{":CCIRC!~guest@freenode-kge.qup.pic9tt.IP MODE CCIRC :+wRix", &IRCMessage{}, false},
		{":CCIRC!~guest@freenode-kge.qup.pic9tt.IP JOIN :#cc", &IRCMessage{}, false},
		{":CCIRC!~guest@freenode-kge.qup.pic9tt.IP PART :#cc", &IRCMessage{}, false},
		{":Guest4454!~guest@freenode-kge.qup.pic9tt.IP NICK :JohnC", &IRCMessage{}, false},
	}

	for i, test := range tests {
		msg := parseIRCMessage(test.in)
		if test.msg.Cmd != msg.Cmd {
			t.Errorf("%d: parseIRCMessage(%q) = %v; want %v", i, test.in, msg, test.msg)
		}
	}

}
