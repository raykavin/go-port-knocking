package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type KnockStep struct {
	Port  int
	Count int
}

var (
	knockSequence = []KnockStep{
		{Port: 7001, Count: 3},
		{Port: 8002, Count: 1},
		{Port: 9003, Count: 2},
	}

	timeout = 1 * time.Second // Max delay for next knocking
)

type ClientState struct {
	StepIndex int
	HitCount  int
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
		log.Fatalf("Error listening on port %d: %v", port, err)
	}
	log.Printf("Listening for knock on port %d", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			if err := conn.Close(); err != nil {
				panic(err)
			}
			continue
		}
		if err := conn.Close(); err != nil {
			panic(err)
		}

		processKnock(ip, port)
	}
}

func processKnock(ip string, port int) {
	mutex.Lock()
	defer mutex.Unlock()

	state, ok := clients[ip]

	// New client or timeout: reset
	if !ok || time.Since(state.LastKnock) > timeout {
		state = &ClientState{}
		clients[ip] = state
	}

	// Extra security
	if state.StepIndex >= len(knockSequence) {
		delete(clients, ip)
		return
	}

	step := knockSequence[state.StepIndex]

	if port == step.Port {
		state.HitCount++
		state.LastKnock = time.Now()

		log.Printf(
			"Knock OK %s | port %d (%d/%d) step %d/%d",
			ip,
			port,
			state.HitCount,
			step.Count,
			state.StepIndex+1,
			len(knockSequence),
		)

		// Knocking complete for this step
		if state.HitCount == step.Count {
			state.StepIndex++
			state.HitCount = 0

			// Complete sequency
			if state.StepIndex == len(knockSequence) {
				log.Printf("ACCESS GRANTED for IP %s", ip)
				delete(clients, ip)

				fmt.Println("OK...")
			}
		}
	} else {
		log.Printf("Invalid knock from %s (port %d, expected %d)",
			ip,
			port,
			step.Port)
		delete(clients, ip)
	}
}

func server() {
	unPorts := make(map[int]struct{})

	for _, step := range knockSequence {
		unPorts[step.Port] = struct{}{}
	}

	for port := range unPorts {
		go handleKnock(port)
	}

	log.Println("Port knocking server running...")
	select {}
}
