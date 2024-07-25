package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type httpResponse struct {
	StatusLine           string
	StatusCode           string
	OptionalReasonPhrase string
	Headers              []byte
	ResponseBody         []byte
}

func NewResponse(status, code, reasonPhrase string, header []byte, resp []byte) *httpResponse {
	return &httpResponse{
		StatusLine:           status,
		StatusCode:           code,
		OptionalReasonPhrase: reasonPhrase,
		Headers:              header,
		ResponseBody:         resp,
	}
}

func (r httpResponse) Response() []byte {
	res := make([]byte, 0)
	res = append(res, []byte(r.StatusLine)...)
	res = append(res, []byte(" ")...)
	res = append(res, []byte(r.StatusCode)...)
	res = append(res, []byte(" ")...)
	res = append(res, []byte(r.OptionalReasonPhrase)...)
	res = append(res, []byte("\r\n")...)
	res = append(res, r.Headers...)
	res = append(res, []byte("\r\n")...)
	res = append(res, r.ResponseBody...)

	return res
}

type RequestLine struct {
	httpMethod    string
	requestTarget string
	httpVersion   string
}

type ResponseHeader struct {
	ContentType   string
	ContentLength string
}

func (h *ResponseHeader) Header() []byte {
	g := []byte{}
	g = append(g, []byte("Content-Type: ")...)
	g = append(g, []byte(h.ContentType)...)
	g = append(g, []byte("\r\n")...)
	g = append(g, []byte("Content-Length: ")...)
	g = append(g, []byte(h.ContentLength)...)
	g = append(g, []byte("\r\n")...)
	return g
}

type RequestHeader struct {
	Host      string
	UserAgent string
	Accept    string
}

type Request struct {
	RequestLine RequestLine
	Headers     RequestHeader
	RequestBody []byte
}

func ParseRequest(raw []byte) (Request, error) {
	var r Request
	data := strings.Split(string(raw), "\r\n")

	requestLine := strings.Fields(string(data[0]))
	r.RequestLine.httpMethod = requestLine[0]
	r.RequestLine.requestTarget = requestLine[1]
	r.RequestLine.httpVersion = requestLine[2]

	//userAgent := string(data[3])
	for _, val := range data[1:] {
		if strings.HasPrefix(val, "Host: ") {
			r.Headers.Host = strings.TrimPrefix(val, "Host: ")
		}
		if strings.HasPrefix(val, "Accept:: ") {
			r.Headers.Accept = strings.TrimPrefix(val, "Accept: ")
		}
		if strings.HasPrefix(val, "User-Agent: ") {
			r.Headers.UserAgent = strings.TrimPrefix(val, "User-Agent: ")
		}
	}

	//r.Headers.Host = strings.Split(header[0], " ")[1]
	//r.Headers.UserAgent = strings.Split(header[1], " ")[1]
	//r.Headers.Accept = strings.Split(header[2], " ")[1]

	return r, nil

}

func isEmptyReq(path string) bool {
	if path == "/" {
		return true
	}
	return false
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	con, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	buffer := make([]byte, 1024)

	bytesRead, err := con.Read(buffer)
	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return
	}
	req, err := ParseRequest(buffer[:bytesRead])

	if err != nil {
		return
	}
	path := req.RequestLine.requestTarget
	if isEmptyReq(path) {
		resp := NewResponse("HTTP/1.1", "200", "OK", []byte{}, []byte{})
		con.Write(resp.Response())
	} else if strings.HasPrefix(path, "/echo") {
		ans := strings.TrimPrefix(path, "/echo/")
		h := ResponseHeader{
			ContentType:   "text/plain",
			ContentLength: strconv.Itoa(len(ans)),
		}
		resp := NewResponse("HTTP/1.1", "200", "OK", h.Header(), []byte(ans))
		con.Write(resp.Response())
	} else if strings.HasPrefix(path, "/user-agent") {
		h := ResponseHeader{
			ContentType:   "text/plain",
			ContentLength: strconv.Itoa(len(req.Headers.UserAgent)),
		}
		resp := NewResponse("HTTP/1.1", "200", "OK", h.Header(), []byte(req.Headers.UserAgent))
		con.Write(resp.Response())
	} else {
		resp := NewResponse("HTTP/1.1", "404", "Not Found", []byte{}, []byte{})
		con.Write(resp.Response())
	}

	con.Close()

}
