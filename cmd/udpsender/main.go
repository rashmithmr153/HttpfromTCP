package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("error:", err)
	}
	conn, err1 := net.DialUDP("udp", addr, addr)
	if err1 != nil {
		log.Fatal("error:", err1)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error:", err)
		}
		conn.Write([]byte(str))
	}
}
