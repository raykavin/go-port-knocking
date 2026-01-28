package main

import "time"

func main() {
	go server()
	time.Sleep(5 * time.Second)
	client()
}
