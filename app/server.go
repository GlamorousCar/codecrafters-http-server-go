package main

import (
	"fmt"
	"net"
	"os"
)

type httpResponse struct {
	StatusLine           string
	StatusCode           string
	OptionalReasonPhrase string
	Headers              []byte
	ResponseBody         []byte
}

func NewResponse(status, code, reasonPhrase string) *httpResponse {
	return &httpResponse{
		StatusLine:           status,
		StatusCode:           code,
		OptionalReasonPhrase: reasonPhrase,
		Headers:              nil,
		ResponseBody:         nil,
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

	resp := NewResponse("HTTP/1.1", "200", "OK")

	con.Write(resp.Response())
	con.Close()

}
