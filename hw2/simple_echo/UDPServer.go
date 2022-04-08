/**
 * UDPServer.go
 **/

package main

import ("bytes"; "fmt"; "net")

func main() {
    serverPort := "33875"

    pconn, _:= net.ListenPacket("udp", ":"+serverPort)
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)

    buffer := make([]byte, 1024)

    for {
        count, r_addr, _:= pconn.ReadFrom(buffer)
        fmt.Printf("UDP message from %s\n", r_addr.String())
        pconn.WriteTo(bytes.ToUpper(buffer[:count]), r_addr)
    }
}

