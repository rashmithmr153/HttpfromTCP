package main

import (
	"Batman/internal/request"
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")

	//data, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error occured", err)
		}
		fmt.Println("Connection accepted sucessfully")
		userRequest, err := request.RequestFromReader(conn)
		fmt.Println("Request-line:")
		fmt.Println("- Method:", userRequest.RequestLine.Method)
		fmt.Println("- Target:", userRequest.RequestLine.RequestTarget)
		fmt.Println("- Version", userRequest.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for k, v := range userRequest.Header {
			fmt.Printf("- %s: %s\n", k, v)
		}

		fmt.Println("Body:")
		fmt.Println(string(userRequest.Body))
		fmt.Print("Connection has been closed")
	}

}
