package main

import (
	"fmt"
	"net"
	"httpfromtcp/internal/request"
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

		req, err := request.RequestFromReader(connection)
		if err != nil {
			fmt.Printf("error reading request sent over the connection")
		}

		fmt.Printf("\n---------------------------------------------\n")
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", string(req.Body))
		fmt.Printf("\n---------------------------------------------\n")
	}
}
