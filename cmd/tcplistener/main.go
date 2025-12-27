package main

import (
	"fmt"
	"log"
	"net"
	"tcp2http/internal/request"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()

			r, err := request.RequestFromReader(conn)
			if err != nil {
				log.Println("bad request:", err)
				return
			}

			fmt.Println("Method:", r.Method)
			fmt.Println("Target:", r.Target)
			fmt.Println("Version:", r.Version)

			fmt.Println("Headers:")
			for name, values := range r.Headers.Canonical() {
				for _, v := range values {
					fmt.Printf("- %s: %s\n", name, v)
				}
			}

			fmt.Println("Body:")
			fmt.Println(string(r.Body))

		}(c)
	}
}
