package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("error creating a tcp listener %s\n", err)
		return
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("error creating a tcp connection %s\n", err)
			return
		}
		defer connection.Close()

		fmt.Printf("established a tcp connection on port 42069\n")
		lineChannel := getLinesChannel(connection)
		for line := range lineChannel {
			fmt.Println(line)
		}
	}
}

func getLinesChannel(connection net.Conn) <-chan string {
	ch := make(chan string)
	go readRoutine(connection, ch)
	return ch 
}

func readRoutine(connection net.Conn, ch chan string) {
	chunks := make([]byte, 8)
	var text string

	for {
		n, err := connection.Read(chunks)

		if n > 0 {
			text += string(chunks[:n])
            lines := strings.Split(text, "\n")
			
			for idx, line := range lines {
				if idx != len(lines) - 1 {
					ch <- line 
				} else {
					text = line
				}
			}
		}

		if err != nil {
			break
		}
	}

	if len(text) > 0 {
		ch <- text
	}

	close(ch)
}
