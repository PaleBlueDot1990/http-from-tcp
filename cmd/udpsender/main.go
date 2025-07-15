package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udp_address, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf("error resolving the udp address: %s", err)
		return
	}

	upd_conn, err := net.DialUDP("udp", nil, udp_address)
	if err != nil {
		fmt.Printf("error establishing the udp connection: %s", err)
	}
	defer upd_conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
        line, err := reader.ReadString(byte('\n'))
		if err != nil {
			fmt.Printf("error reading the buffer from os stdin")
			return
		}

		_, err = upd_conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("error writing the string to udp connection")
			return
		}
	}
}
