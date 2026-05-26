package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
)

type HTTPStatus int

const (
	StatusOk       HTTPStatus = 200
	StatusNotFound HTTPStatus = 404
)

const (
	AppFolder      = "app/"
	RequestNewLine = "\r\n"
)

func main() {
	InitiPages()
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error listening", err)
	}
	fmt.Println("Listening on port 8090 ")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	for {

		buffer := make([]byte, 1024)

		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("Error reading or connection lost")
			break
		}

		request := parseRequest(string(buffer[:n]))

		fmt.Printf("Received %d bytes: %s\n", n, string(buffer[:n]))

		// Handle POST request
		if request.method == "POST" {
			escapedName := url.QueryEscape(request.body["name"])
			conn.Write([]byte(httpRedirect(request.path + fmt.Sprintf("?name=%s", escapedName))))
			return
		}

		// Handle GET request
		if request.method == "GET" {

			filePath := request.path
			if strings.Contains(filePath, "..") {
				conn.Write([]byte(httpResponse(StatusNotFound, "text/html", "<h1>Page Not Found</h1>")))
				return
			}

			// Handle html in routes
			handler, ok := Routes[request.path]
			if ok {
				responseString := handler(request)
				conn.Write([]byte(httpResponse(StatusOk, "text/html", responseString)))
			}

			// Fallback to static folder
			// Need to match path to the filesystem
			// Read the file, catch the error if need be
			targetFile := AppFolder + filePath
			if _, err := os.Stat(targetFile); os.IsNotExist(err) {
				// If the raw path doesn't exist, try appending html
				if _, err := os.Stat(targetFile + ".html"); err == nil {
					targetFile = targetFile + ".html"
				}
			}
			fileBytes, err := os.ReadFile(targetFile)
			if err != nil {
				conn.Write([]byte(httpResponse(StatusNotFound, "text/html", "<h1>Page Not Found</h1>")))
				return
			}

			status := StatusOk
			responseString := string(fileBytes)
			// Determine the correct Content-Type for the static asset
			contentType := "text/plain"
			if strings.HasSuffix(request.path, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(request.path, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(request.path, ".html") {
				contentType = "text/html"
			}

			conn.Write([]byte(httpResponse(status, contentType, responseString)))
		}
	}
}

type HTTPRequest struct {
	method      string
	path        string
	body        map[string]string
	queryParams map[string]string
}

func parseRequest(request string) HTTPRequest {
	// 1. Isolate the body form the headers
	// We split into exactly 2 parts [0] is headers, [1] is the body payload

	parts := strings.SplitN(request, "\r\n\r\n", 2)

	if len(parts) < 2 {
		parts = strings.SplitN(request, "\n\n", 2)
	}

	headerChunk := parts[0]
	bodyChunk := ""
	if len(parts) > 1 {
		bodyChunk = strings.TrimSpace(parts[1])
	}

	// Need to only parse the first line of the request
	// First line is in the format METHOD PATH VERSION
	firstLine := strings.SplitN(headerChunk, "\n", 2)[0]
	firstLine = strings.TrimSpace(firstLine)
	s := strings.Split(firstLine, " ")

	//  Safety Check: Ensure we actually got a method and a path!
	if len(s) < 2 {
		// Return a safe default instead of crashing
		return HTTPRequest{method: "GET", path: "/"}
	}

	path := s[1]
	var queryParams map[string]string
	var body map[string]string

	qps := strings.SplitN(path, "?", 2)
	route := qps[0]
	if len(qps) > 1 {
		queryParams = parseParams(qps[1])
	} else {
		queryParams = make(map[string]string)
	}

	if s[0] == "POST" {
		body = parseParams(bodyChunk)
	} else {
		body = make(map[string]string)
	}

	return HTTPRequest{s[0], route, body, queryParams}
}

func parseParams(qps string) map[string]string {
	pairs := strings.Split(qps, "&")
	params := make(map[string]string, len(pairs))

	for _, pair := range pairs {
		param := strings.Split(pair, "=")
		if len(param) == 2 {
			decodedValue, err := url.QueryUnescape(param[1])
			if err != nil {
				// Fallback
				decodedValue = param[1]
			}
			params[param[0]] = decodedValue
		}
	}
	return params
}

func httpResponse(status HTTPStatus, contentType string, message string) string {
	length := len(message)
	return fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: %s\r\nContent-Length:%d\r\n\r\n%s", status, contentType, length, message)
}

func httpHTMLResponse(status HTTPStatus, message string) string {
	return httpResponse(status, "text/html", message)
}

func replaceKeyInString(s string, key string, value string) string {
	return strings.ReplaceAll(s, fmt.Sprintf("{{%s}}", key), value)
}

func httpRedirect(location string) string {
	return fmt.Sprintf("HTTP/1.1 303 See Other\r\nLocation: %s\r\nContent-Length: 0\r\n\r\n", location)
}
