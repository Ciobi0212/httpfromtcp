package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Panic(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Panic(err)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		input, err := reader.ReadBytes('\n')
		if err != nil {
			log.Panic(err)
		}

		_, err = conn.Write(input)
		if err != nil {
			log.Panic(err)
		}
	}
}
