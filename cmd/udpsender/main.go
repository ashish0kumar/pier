package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}

		_, _ = conn.Write([]byte(line))
	}
}
