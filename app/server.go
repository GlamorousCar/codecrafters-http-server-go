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
	Headers              ResponseHeader
	ResponseBody         []byte
}

//	func NewResponse(status, code, reasonPhrase string, header []byte, resp []byte) *httpResponse {
//		return &httpResponse{
//			StatusLine:           status,
//			StatusCode:           code,
//			OptionalReasonPhrase: reasonPhrase,
//			Headers:              header,
//			ResponseBody:         resp,
//		}
//	}
func (r httpResponse) Response() []byte {
	res := make([]byte, 0)
	res = append(res, []byte(r.StatusLine)...)
	res = append(res, []byte(" ")...)
	res = append(res, []byte(r.StatusCode)...)
	res = append(res, []byte(" ")...)
	res = append(res, []byte(r.OptionalReasonPhrase)...)
	res = append(res, []byte("\r\n")...)
	res = append(res, []byte("Content-Encoding: ")...)
	res = append(res, []byte(r.Headers.ContentEncoding)...)
	res = append(res, []byte("\r\n")...)
	res = append(res, []byte("Content-Type: ")...)
	res = append(res, []byte(r.Headers.ContentType)...)
	res = append(res, []byte("\r\n")...)
	res = append(res, []byte("Content-Length: ")...)
	res = append(res, []byte(r.Headers.ContentLength)...)
	res = append(res, []byte("\r\n")...)
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
	ContentType     string
	ContentLength   string
	ContentEncoding string
}

func (h ResponseHeader) ToBytes() []byte {
	g := []byte{}

	return g
}

type RequestHeader struct {
	Host           string
	UserAgent      string
	Accept         string
	AcceptEncoding []string
}

type Request struct {
	RequestLine RequestLine
	Headers     RequestHeader
	RequestBody []byte
}

func (r *Request) parseData(reqString string) {
	switch msg := reqString; {
	case strings.HasPrefix(msg, "GET") || strings.HasPrefix(msg, "POST"):
		data := strings.Fields(msg)
		r.RequestLine.httpMethod = data[0]
		r.RequestLine.requestTarget = data[1]
		r.RequestLine.httpVersion = data[2]

	case strings.HasPrefix(msg, "Host"):
		r.Headers.Host = strings.Fields(msg)[1]

	case strings.HasPrefix(msg, "Accept:"):
		r.Headers.Accept = strings.Fields(msg)[1]

	case strings.HasPrefix(msg, "User-Agent"):
		r.Headers.UserAgent = strings.Fields(msg)[1]

	case strings.HasPrefix(msg, "Accept-Encoding:"):
		r.Headers.AcceptEncoding = strings.Fields(msg)[1:]
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
	fmt.Println(r)
	//req, err := ParseRequest()
	var resp httpResponse
	switch urlPath := r.RequestLine.requestTarget; {
	case urlPath == "/":
		resp = httpResponse{
			StatusLine:           "HTTP/1.1",
			StatusCode:           "200",
			OptionalReasonPhrase: "OK",
		}

	case strings.HasPrefix(urlPath, "/echo"):

		ans := strings.TrimPrefix(urlPath, "/echo/")
		resp = httpResponse{
			StatusLine:           "HTTP/1.1",
			StatusCode:           "200",
			OptionalReasonPhrase: "OK",
			Headers: ResponseHeader{
				ContentType:   "text/plain",
				ContentLength: strconv.Itoa(len(ans)),
			},
			ResponseBody: []byte(ans),
		}

	case strings.HasPrefix(urlPath, "/user-agent"):

		resp = httpResponse{
			StatusLine:           "HTTP/1.1",
			StatusCode:           "200",
			OptionalReasonPhrase: "OK",
			Headers: ResponseHeader{
				ContentType:   "text/plain",
				ContentLength: strconv.Itoa(len(r.Headers.UserAgent)),
			},
			ResponseBody: []byte(r.Headers.UserAgent),
		}
	case strings.HasPrefix(urlPath, "/files/"):

		switch r.RequestLine.httpMethod {
		case "GET":
			filename := strings.TrimPrefix(urlPath, "/files/")
			filePath := path.Join(dir, filename)
			if _, err = os.Stat(filePath); os.IsNotExist(err) {
				resp = httpResponse{
					StatusLine:           "HTTP/1.1",
					StatusCode:           "404",
					OptionalReasonPhrase: "Not Found",
					Headers:              ResponseHeader{},
					ResponseBody:         nil,
				}
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
				resp = httpResponse{
					StatusLine:           "HTTP/1.1",
					StatusCode:           "200",
					OptionalReasonPhrase: "OK",
					Headers: ResponseHeader{
						ContentType:   "application/octet-stream",
						ContentLength: strconv.Itoa(n),
					},
					ResponseBody: data[:n],
				}

			}
		case "POST":
			filename := strings.TrimPrefix(urlPath, "/files/")
			filePath := path.Join(dir, filename)
			os.WriteFile(filePath, r.RequestBody, 0777)

			resp = httpResponse{
				StatusLine:           "HTTP/1.1",
				StatusCode:           "201",
				OptionalReasonPhrase: "Created",
				Headers:              ResponseHeader{},
				ResponseBody:         []byte{},
			}

		}

	default:
		resp = httpResponse{
			StatusLine:           "HTTP/1.1",
			StatusCode:           "404",
			OptionalReasonPhrase: "Not Found",
			Headers:              ResponseHeader{},
			ResponseBody:         []byte{},
		}
	}
	for _, val := range r.Headers.AcceptEncoding {
		if val == "gzip" {
			resp.Headers.ContentEncoding = "gzip"
		}
	}

	conn.Write(resp.Response())
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
