package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Printf("error opening file: %s", err)
		return 
	}
	defer file.Close()

	lineChannel := getLinesChannel(file)
	for line := range lineChannel {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	ch := make(chan string)
	go readRoutine(file, ch)
	return ch 
}

func readRoutine(file io.ReadCloser, ch chan string) {
	chunks := make([]byte, 8)
	var text string

	for {
		n, err := file.Read(chunks)

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
