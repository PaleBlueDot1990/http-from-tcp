package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine 
	Headers     headers.Headers 
	Body        []byte 
	ParserState string 
}

type RequestLine struct {
	HttpVersion   string 
	RequestTarget string 
	Method        string 
}

/*
 It's important to understand the difference. When we read, all we're doing is moving the data 
 from the reader (which in the case of HTTP is a network connection, but it could be a file as
 well, our code is agnostic) into our program. When we parse, we're taking that data and 
 interpreting it (moving it from a []byte to a RequestLine struct). Once its parsed, we can 
 discard it from the buffer to save memory.
*/
func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	request := Request{
		ParserState: "PARSE_METHOD",
		Headers:     make(headers.Headers),
		Body:        make([]byte, 0),
	}

	for request.ParserState != "DONE" {
		if readToIndex == cap(buf) {
			newBuf := make([]byte, 2*cap(buf))
			copy(newBuf, buf)
			buf = newBuf 
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			bytesParsed, err := request.parse(buf[:readToIndex])
			if err != nil {
				return nil, err 
			}
			if bytesParsed == 0 && (request.ParserState == "PARSE_METHOD" || request.ParserState == "PARSE_HEADERS") {
				return nil, fmt.Errorf("incomplete http request")
			}
			if request.ParserState == "PARSE_BODY" {
				val, ok := request.Headers["content-length"]
				if ok {
					contentLength, _ := strconv.Atoi(val)
					if len(request.Body) < contentLength {
						return nil, fmt.Errorf("incomplete http request")
					}
				}
			}
			return &request, nil 
		} 
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err 
		}
		if bytesParsed != 0 {
			newBuf := make([]byte, cap(buf))
			copy(newBuf, buf[bytesParsed:])
			buf = newBuf
			readToIndex -= bytesParsed
		}
	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.ParserState {
		case "PARSE_METHOD":
			bytesParsed, err := r.parseRequestLine(string(data))
			if err != nil {
				return -1, err
			} else if bytesParsed == 0 {
				return 0, nil
			} else {
				r.ParserState = "PARSE_HEADERS"
				return bytesParsed, nil
			}
		case "PARSE_HEADERS":
			bytesParsed, done, err := r.Headers.Parse(data)
			if err != nil {
				return -1, err 
			} else if bytesParsed == 0 {
				return 0, nil 
			} else if !done {
				return bytesParsed, nil
			} else {
				r.ParserState = "PARSE_BODY"
				return bytesParsed, nil 
			}
		case "PARSE_BODY":
			val, ok := r.Headers["content-length"]
			if !ok {
				r.ParserState = "DONE"
				return 0, nil 
			}
			newLength := len(r.Body)+len(data)
			r.Body = append(r.Body, data...)
			r.Body = r.Body[:newLength]

			contentLength, _ := strconv.Atoi(val)
			if len(r.Body) > contentLength {
				return -1, fmt.Errorf("body length exceeds content-length limit of %d", contentLength)
			}
			if len(r.Body) == contentLength {
				r.ParserState = "DONE"
				return len(data), nil 
			}
			return len(data), nil 
		default:
			return -1, fmt.Errorf("unknown state")
	}
}

func (r *Request) parseRequestLine(line string) (int, error) {
	clrfIndex := strings.Index(line, "\r\n")
	if clrfIndex == -1 {
		return 0, nil
	}

	parts1 := strings.Split(line, "\r\n")[0]
	parts2 := strings.Split(parts1, " ")
	if len(parts2) != 3 {
		return -1, fmt.Errorf("poorly formatted request line")
	}

	method := parts2[0]
	if !isMethodCorrect(method) {
		return -1, fmt.Errorf("method name is incorrect")
	}

	requestTarget := parts2[1]
	if !isTargetCorrect(requestTarget) {
		return -1, fmt.Errorf("request target is incorrect")
	}

	httpVersion := parts2[2]
	if !isHttpVersionCorrect(httpVersion) {
		return -1, fmt.Errorf("http version is incorrect")
	}

	requestLine := RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   strings.Split(httpVersion, "/")[1],
	}

	r.RequestLine = requestLine
	return clrfIndex + 2, nil
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
