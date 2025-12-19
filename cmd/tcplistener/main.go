package main

import (
	"fmt"
	"log"
	"net"
	"tcp2http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", r.Body)
	}
}
