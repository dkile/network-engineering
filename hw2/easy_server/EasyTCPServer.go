/**
 * TCPServer.go
 **/

package main

import (
	"bytes"
	"fmt"
	"net"
)

func main() {
    serverPort := "33875"

    listener, _:= net.Listen("tcp", ":" + serverPort)
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)

    buffer := make([]byte, 1024)

    for {
        conn, _:= listener.Accept()
        fmt.Printf("Connection request from %s\n", conn.RemoteAddr().String())
        count, _ := conn.Read(buffer)
        conn.Write(bytes.ToUpper(buffer[:count]))
        conn.Close()
    }
}

