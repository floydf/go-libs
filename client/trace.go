package client

import (
	"fmt"
	"net/http/httptrace"
)

var trace = &httptrace.ClientTrace{
	ConnectStart: func(network, addr string) {
		fmt.Printf("Connection Start: network=%q addr=%q\n", network, addr)
	},

	ConnectDone: func(network, addr string, err error) {
		fmt.Printf("Connection Done: network=%q addr=%q err=%+v\n", network, addr, err)
	},

	GotConn: func(connInfo httptrace.GotConnInfo) {
		fmt.Printf("Got Conn: %+v\n", connInfo)
	},
	DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
		fmt.Printf("DNS Info: %+v\n", dnsInfo)
	},

	PutIdleConn: func(err error) {
		fmt.Printf("PutIdleConn: err=%+v\n", err)
	},
}
