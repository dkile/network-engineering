/**
 * UDPClient.go
 * Name: Jeong Yong Jun
 * StudentID: 20173875
 **/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type message struct {
	Option  int `json:"option"`
	Content string `json:"content"`
}

func main() {

	openSignalChannel("Bye bye~")

	// define server name, port
	serverName := "localhost"
	serverPort := "23875"

	server_addr, _ := net.ResolveUDPAddr("udp", serverName+":"+serverPort)

	// define std input reader
	std_reader := bufio.NewReader(os.Stdin)

	// create UDP socket
	pconn, _ := net.ListenPacket("udp", ":")

	localAddr := pconn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("Client is running on port %d\n", localAddr.Port)

	defer pconn.Close()

	for {
		// choose input option
		fmt.Printf("<Menu>\n" +
						"1) convert text to UPPER-case\n" +
						"2) get my IP address and port number\n" +
            "3) get server request count\n" +
            "4) get server running time\n" +
            "5) exit\n" +
            "Input option: ")
		input_option, _ := std_reader.ReadString('\n')
		option, _ := strconv.Atoi(strings.Trim(input_option, "\n"))

		switch option {
		case 1:
			fmt.Printf("Input lowercase sentence: ")
			input_sentence, _ := std_reader.ReadString('\n')
			req_msg := createMessage(1, strings.Trim(input_sentence, "\n"))
			sendRequest(&pconn, server_addr, req_msg)
			res := getResponse(&pconn)
			fmt.Printf("Reply from server: %s\n", res.Content)
		case 2:
			req_msg := createMessage(option, "")
			sendRequest(&pconn, server_addr, req_msg)
			res := getResponse(&pconn)
			ip := strings.Split(res.Content, ":")[0]
			port := strings.Split(res.Content, ":")[1]
			fmt.Printf("Reply from server: client IP=%s, port=%s\n", ip, port)
		case 3:
			req_msg := createMessage(option, "")
			sendRequest(&pconn, server_addr, req_msg)
			res := getResponse(&pconn)
			fmt.Printf("Reply from server: requests served=%s\n", res.Content)
		case 4:
			req_msg := createMessage(option, "")
			sendRequest(&pconn, server_addr, req_msg)
			res := getResponse(&pconn)
			fmt.Printf("Reply from server: runtime=%s\n", res.Content)
		case 5:
			pconn.Close()
			fmt.Printf("Bye bye~\n")
			os.Exit(0)
		default:
			fmt.Printf("Please enter number between 1 and 5\n")
		}
	}
}

func openSignalChannel(ment string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		sig := <-sigs
		if sig.String() == "interrupt" {
			fmt.Printf("\n%s\n", ment)
		} else {
			fmt.Printf("\nunhandled signal\n")
		}
		os.Exit(0)
	}()
}

func createMessage(opt int, content string) message {
	req_msg := message{Option: opt, Content: content}
	return req_msg
}

/**
*
 */
func sendRequest(conn *net.PacketConn, server_addr *net.UDPAddr, req_msg message) {

	// encode struct to json byte array
	json_msg, err := json.Marshal(req_msg)
	if err != nil {
		log.Fatal("encode error: ", err)
	}

	// send request
	(*conn).WriteTo(json_msg, server_addr)
}

/**
*
 */
func getResponse(conn *net.PacketConn) message {
	var res_msg message
	buffer := make([]byte, 1024)
	count, _, _ := (*conn).ReadFrom(buffer)
	json.Unmarshal(buffer[:count], &res_msg)

	return res_msg
}
