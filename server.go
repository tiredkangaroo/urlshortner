package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// NotFound represents the []byte for "<h1> Not Found </h1> <pre>This redirection link does not exist.</pre>".
var NotFound = [147]byte{72, 84, 84, 80, 47, 49, 46, 49, 32, 52, 48, 52, 32, 78, 111, 116, 32, 70, 111, 117, 110, 100, 92, 114, 92, 110, 67, 111, 110, 116, 101, 110, 116, 45, 84, 121, 112, 101, 58, 116, 101, 120, 116, 47, 104, 116, 109, 108, 92, 114, 92, 110, 67, 111, 110, 116, 101, 110, 116, 45, 76, 101, 110, 103, 116, 104, 58, 32, 54, 57, 92, 114, 92, 110, 92, 114, 92, 110, 60, 104, 49, 62, 32, 78, 111, 116, 32, 70, 111, 117, 110, 100, 32, 60, 47, 104, 49, 62, 32, 60, 112, 114, 101, 62, 84, 104, 105, 115, 32, 114, 101, 100, 105, 114, 101, 99, 116, 105, 111, 110, 32, 108, 105, 110, 107, 32, 100, 111, 101, 115, 32, 110, 111, 116, 32, 101, 120, 105, 115, 116, 46, 60, 47, 112, 114, 101, 62}

// ServerIP provides the local server's IP addr.
var ServerIP = getLocalIP().String()

type redirectionlinkstype = map[string]string

// serverList returns the bytes necessary that provides a list of all
// the possible redirection links.
func serverList(links redirectionlinkstype) []byte {
	li := ""

	for short, long := range links {
		li += fmt.Sprintf(`<li>%s: <a href="%s">%s</a></li><br>`, short, long, long)
	}

	data := fmt.Sprintf(`<p>Version: %s</p><ul>%s</ul>`, version, li)
	return []byte(fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s", len(data), data))
}

// serverRedirect returns the bytes necessary to redirect the client.
// If the client attempted a path that does not exist, it will instead
// return the not found page.
func serverRedirect(links redirectionlinkstype, path string) []byte {
	s := strings.Split(path, "/")
	short := s[1]

	long, ok := links[short]
	if !ok {
		return NotFound[:]
	}
	if len(s) > 2 {
		long += "/" + strings.Join(s[2:], "/")
	}

	return []byte(fmt.Sprintf(
		"HTTP/1.1 301 Moved Permanently\r\n"+
			"Content-Type: text/html\r\n"+
			"Content-Length: 0\r\n"+
			"Location: %s\r\n"+
			"\r\n", long))
}

// handleConnection provides a mean to route HTTP requests
// through a connection.
func handleConnection(conn net.Conn, links *redirectionlinkstype) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	reqline, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Split(reqline, " ")
	if len(parts) < 3 { // invalid request, you don't deserve my processing power
		return
	}

	path := parts[1]
	switch path {
	case "/list":
		conn.Write(serverList(*links))
	default:
		conn.Write(serverRedirect(*links, path))
	}
}

// startServer starts the FastHTTP server at port 80.
func startServer(links *redirectionlinkstype) error {
	var err error
	listener, err := net.Listen("tcp", ":8010")
	if err != nil {
		return err
	}

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err != nil {
			break
		}
		go handleConnection(conn, links)
	}

	return err
}
