/**
 * TCPClient.go
 **/

package main

import ("bufio"; "fmt"; "net"; "os")

func main() {

    serverName := "nsl2.cau.ac.kr"
    serverPort := "12000"

    conn, _:= net.Dial("tcp", serverName+":"+serverPort)

    localAddr := conn.LocalAddr().(*net.TCPAddr)
    fmt.Printf("Client is running on port %d\n", localAddr.Port)

    fmt.Printf("Input lowercase sentence: ")
    input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    conn.Write([]byte(input))

    buffer := make([]byte, 1024)
    conn.Read(buffer)
    fmt.Printf("Reply from server: %s", string(buffer))

    conn.Close()
}
