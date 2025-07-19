package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine 
}

type RequestLine struct {
	HttpVersion string 
	RequestTarget string 
	Method string 
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	line, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return parseRequestLine(string(line))
}

/*
HTTP message format according to RFC 9112:
start-line CRLF              -> request (for request) or response status (for response)
*( field-line CRLF )         -> http headers (key value pairs)
CRLF
[ message-body ]             -> the message body is optional 

We will be assuming that at least one header is present in the message.

HTTP message examples-
GET / HTTP/1.1\r\n
Host: localhost:42069\r\n
\r\n
*/

func parseRequestLine(line string) (*Request, error) {
	clrfIndex := strings.Index(line, "\r\n")
	if clrfIndex == -1 {
		fmt.Printf("request line does not contain clrf")
		return nil, fmt.Errorf("request line does not contain clrf")
	}

	parts1 := strings.Split(line, "\r\n")[0]
	parts2 := strings.Split(parts1, " ")
	if len(parts2) != 3 {
		fmt.Printf("poorly formatted request line")
		return nil, fmt.Errorf("poorly formatted request line")
	}

	method := parts2[0]
	if !isMethodCorrect(method) {
		fmt.Printf("method name is incorrect")
		return nil, fmt.Errorf("method name is incorrect")
	}

	requestTarget := parts2[1]
	if !isTargetCorrect(requestTarget) {
		fmt.Printf("request target is incorrect")
		return nil, fmt.Errorf("request target is incorrect")
	}

	httpVersion := parts2[2]
	if !isHttpVersionCorrect(httpVersion) {
		fmt.Printf("http version is incorrect")
		return nil, fmt.Errorf("http version is incorrect")
	}

	requestLine := RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   strings.Split(httpVersion, "/")[1],
	}

	request := Request {
		RequestLine: requestLine,
	}

	return &request, nil
}

func isMethodCorrect(method string) bool {
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
		return false
	}
	return true 
}

func isTargetCorrect(target string) bool {
	if len(target) == 0 {
		return false
	}

	if len(target) == 1 {
		return target == "/"
	}

	if target[0] != '/' {
		return false
	}

	parts := strings.Split(target[1:], "/")
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
	}

	return true
}

func isHttpVersionCorrect(httpVersion string) bool {
	cnt := strings.Count(httpVersion, "/")
	if cnt != 1 {
		return false
	}

	parts := strings.Split(httpVersion, "/")

	protocol := parts[0]
	if protocol != "HTTP" {
		return false
	}

	return parts[1] == "1.1" 
}
