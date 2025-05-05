package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Ciobi0212/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Panic(err)
	}

	defer listener.Close()

	for {
		log.Println("Listening for connection")

		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}

		log.Println("Connection received")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Body: ")
		fmt.Println(string(req.Body))
	}

}
