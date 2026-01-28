package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	knockSequence = []int{7001, 8002, 9003}
	timeout       = 5 * time.Second
)

type ClientState struct {
	Index     int
	LastKnock time.Time
}

var (
	clients = make(map[string]*ClientState)
	mutex   sync.Mutex
)

func handleKnock(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error listening %d: %v", port, err)
	}
	log.Printf("Listening knock on port: %d", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

		if err = conn.Close(); err != nil {
			panic(err)
		}

		processKnock(ip, port)
	}
}

func processKnock(ip string, port int) {
	mutex.Lock()
	defer mutex.Unlock()

	state, ok := clients[ip]
	if !ok || time.Since(state.LastKnock) > timeout {
		state = &ClientState{}
		clients[ip] = state
	}

	expectedPort := knockSequence[state.Index]
	if port == expectedPort {
		state.Index++
		state.LastKnock = time.Now()

		if state.Index == len(knockSequence) {
			log.Printf("Knock sequency ok for IP: %s | ACCESS GRANTED", ip)
			delete(clients, ip)

			fmt.Println("ok...")
		}
	} else {
		log.Printf("Invalid knock sequency of %s (port %d)", ip, port)
		delete(clients, ip)
	}
}

func server() {
	for _, port := range knockSequence {
		go handleKnock(port)
	}

	select {}
}
