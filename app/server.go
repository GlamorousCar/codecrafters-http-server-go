package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
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

func (r *Request) parseData(reqString string) {
	switch msg := reqString; {
	case strings.HasPrefix(msg, "GET"):
		data := strings.Fields(msg)
		r.RequestLine.httpMethod = data[0]
		r.RequestLine.requestTarget = data[1]
		r.RequestLine.httpVersion = data[2]

	case strings.HasPrefix(msg, "Host"):
		r.Headers.Host = strings.Fields(msg)[1]

	case strings.HasPrefix(msg, "Accept"):
		r.Headers.Accept = strings.Fields(msg)[1]

	case strings.HasPrefix(msg, "User-Agent"):
		r.Headers.UserAgent = strings.Fields(msg)[1]
	default:
		r.RequestBody = []byte(reqString)
	}
}

func handleConn(conn net.Conn) {
	r := &Request{}
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return
	}
	for _, val := range strings.Split(string(buffer[:bytesRead]), "\n") {
		r.parseData(val)
	}

	//req, err := ParseRequest()

	switch urlPath := r.RequestLine.requestTarget; {
	case urlPath == "/":
		resp := NewResponse("HTTP/1.1", "200", "OK", []byte{}, []byte{})
		conn.Write(resp.Response())

	case strings.HasPrefix(urlPath, "/echo"):

		ans := strings.TrimPrefix(urlPath, "/echo/")
		h := ResponseHeader{
			ContentType:   "text/plain",
			ContentLength: strconv.Itoa(len(ans)),
		}
		resp := NewResponse("HTTP/1.1", "200", "OK", h.Header(), []byte(ans))
		conn.Write(resp.Response())

	case strings.HasPrefix(urlPath, "/user-agent"):
		h := ResponseHeader{
			ContentType:   "text/plain",
			ContentLength: strconv.Itoa(len(r.Headers.UserAgent)),
		}
		resp := NewResponse("HTTP/1.1", "200", "OK", h.Header(), []byte(r.Headers.UserAgent))
		conn.Write(resp.Response())
	case strings.HasPrefix(urlPath, "/files/"):
		filename := strings.TrimPrefix(urlPath, "/files/")
		filePath := path.Join(dir, filename)
		if _, err = os.Stat(filePath); os.IsNotExist(err) {
			resp := NewResponse("HTTP/1.1", "404", "Not Found", []byte{}, []byte{})
			conn.Write(resp.Response())
		} else {
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println("Unable to open file:", err)
				return
			}
			defer file.Close()

			reader := bufio.NewReader(file)
			data := make([]byte, 1024)
			n, err := reader.Read(data)

			h := ResponseHeader{
				ContentType:   "application/octet-stream",
				ContentLength: strconv.Itoa(n),
			}
			resp := NewResponse("HTTP/1.1", "200", "OK", h.Header(), []byte(data[:n]))
			conn.Write(resp.Response())
		}

	default:
		resp := NewResponse("HTTP/1.1", "404", "Not Found", []byte{}, []byte{})
		conn.Write(resp.Response())
	}

	//fmt.Println(err)

	//scanner := bufio.NewScanner(reader)
	//scanner.Split(ScanCRLF)
	//fmt.Println(scanner.Text())
	//fmt.Println(scanner.Text())
	//fmt.Println(scanner.Text())
	//fmt.Println(scanner.Text())
	//req := Request{}
	//for scanner.Scan() {
	//	line := scanner.Text()
	//
	//	fmt.Println("---", line)
	//}
	//if err := scanner.Err(); err != nil {
	//	fmt.Printf("Invalid input: %s", err)
	//}
	//fmt.Println("==", req)
	//
	//if err != nil {
	//	// Handle error or end of connection
	//	break
	//}
	//fmt.Print("Received: ", request)

	//req, err := ParseRequest()
	//if err != nil {
	//	return
	//}
	//

	defer conn.Close()

}

var dir string

func main() {

	flag.StringVar(&dir, "directory", "", "")
	flag.Parse()
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	// Uncomment this block to pass the first stage
	//
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	defer ln.Close()

	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConn(conn)

	}

}
