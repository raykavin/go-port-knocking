package main

import (
	"fmt"
	"net"
	"time"
)

func knock(host string, port int) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err == nil {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}
}

func client() {
	serverIP := "127.0.0.1" // Server address
	knockSeqPorts := []int{7001, 7001, 7001, 8002, 9003, 9003}

	for _, port := range knockSeqPorts {
		knock(serverIP, port)
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Port knocking send")
}
