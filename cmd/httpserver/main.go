package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	t := req.RequestLine.RequestTarget
	if t == "/yourproblem" {
		handler400(w)
		return 
	}
	if t == "/myproblem" {
		handler500(w)
		return
	}
	if strings.HasPrefix(t, "/httpbin/") {
		proxyHandler(req, w)
		return 
	}
	handler200(w)
}

func handler400(w *response.Writer) {
	w.WriteRequestLine(response.StatusBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer) {
	w.WriteRequestLine(response.StatusInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer) {
	w.WriteRequestLine(response.StatusOK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func proxyHandler(req *request.Request, w *response.Writer) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Printf("Proxying to %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		handler500(w)
		return
	}
	defer resp.Body.Close()

	w.WriteRequestLine(response.StatusOK)

	h := response.GetDefaultHeaders(0)
	delete(h, "Content-Length")
	h["Transfer-Encoding"] = "chunked"
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	allChunks := make([]byte, 0)
	for {
		n, err := resp.Body.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading response body: %s", err)
			break
		}

		fmt.Printf("Read %d bytes from %s\n", n, url)
		_, err = w.WriteChunkedBody(buffer[:n])
		if err != nil {
			fmt.Printf("Error writing chunked body to response: %s\n", err)
			break
		}
		fmt.Printf("Writing %d bytes to response\n", n)

		len := len(allChunks) + n 
		allChunks = append(allChunks, buffer[:n]...)
		allChunks = allChunks[:len]
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("Error writing chunked body done: %s", err)
	}
	fmt.Printf("Succesfully wrote all chunked data to response\n")

	hashAllChunks := sha256.Sum256(allChunks)
	hexStr := hex.EncodeToString(hashAllChunks[:])

	trailers := make(headers.Headers)
	trailers["X-Content-SHA256"] = hexStr
	trailers["X-Content-Length"] = fmt.Sprintf("%d", len(allChunks))
	w.WriteTrailers(trailers)
}

