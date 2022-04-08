/**
 * UDPClient.go
 * Name: Jeong Yong Jun
 * StudentID: 20173875
 **/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type message struct {
	Option  int    `json:"option"`
	Content string `json:"content"`
}

func main() {

	openSignalChannel("Bye bye~")

	// define server name, port
	serverName := "localhost"
	serverPort := "23875"

	server_addr, _ := net.ResolveUDPAddr("udp", serverName+":"+serverPort)

	// create UDP socket
	pconn, _ := net.ListenPacket("udp", ":")

	localAddr := pconn.LocalAddr().(*net.UDPAddr)
	fmt.Println("Client is running on port", localAddr.Port)

	defer pconn.Close()

	for {

		// choose input option
		var input_option string
		fmt.Println("<Menu>")
		fmt.Println("1) convert text to UPPER-case")
		fmt.Println("2) get my IP address and port number")
		fmt.Println("3) get server request count")
		fmt.Println("4) get server running time")
		fmt.Println("5) exit")
		fmt.Printf("Input option: ")
		_, err := fmt.Scanln(&input_option)
		if err != nil {
			log.Fatal("scanning input error: ", err)
		}
		option, _ := strconv.Atoi(input_option)

		switch option {
		case 1:
			var input_sentence string
			fmt.Printf("Input lowercase sentence: ")
			fmt.Scanln(&input_sentence)

			req_msg := createJsonMessage(1, strings.Trim(input_sentence, "\n"))

			req_time := time.Now()

			pconn.WriteTo(req_msg, server_addr)

			buffer := make([]byte, 1024)
			count, _, err := pconn.ReadFrom(buffer)

			rtt := float64(time.Now().Sub(req_time)) / float64(time.Millisecond)

			if err != nil {
				log.Fatal("reading response error: ", err)
			}

			_, content := decodeJsonMessage(buffer[:count])

			fmt.Println("Reply from server: ", content)
			fmt.Println("RTT = ", rtt, "ms")
		case 2:
			req_msg := createJsonMessage(option, "")
			req_time := time.Now()
			pconn.WriteTo(req_msg, server_addr)

			buffer := make([]byte, 1024)
			count, _, err := pconn.ReadFrom(buffer)
			rtt := float64(time.Now().Sub(req_time)) / float64(time.Millisecond)

			if err != nil {
				log.Fatal("reading response error: ", err)
			}

			_, content := decodeJsonMessage(buffer[:count])

			ip := strings.Split(content, ":")[0]
			port := strings.Split(content, ":")[1]
			fmt.Println("Reply from server: client IP=", ip, "port=", port)
			fmt.Println("RTT = ", rtt, "ms")
		case 3:
			req_msg := createJsonMessage(option, "")

			req_time := time.Now()

			pconn.WriteTo(req_msg, server_addr)

			buffer := make([]byte, 1024)
			count, _, err := pconn.ReadFrom(buffer)

			rtt := float64(time.Now().Sub(req_time)) / float64(time.Millisecond)

			if err != nil {
				log.Fatal("reading response error: ", err)
			}

			_, content := decodeJsonMessage(buffer[:count])
			fmt.Println("Reply from server: requests served=", content)
			fmt.Println("RTT = ", rtt, "ms")
		case 4:
			req_msg := createJsonMessage(option, "")

			req_time := time.Now()

			pconn.WriteTo(req_msg, server_addr)

			buffer := make([]byte, 1024)
			count, _, err := pconn.ReadFrom(buffer)

			rtt := float64(time.Now().Sub(req_time)) / float64(time.Millisecond)

			if err != nil {
				log.Fatal("reading response error: ", err)
			}

			_, content := decodeJsonMessage(buffer[:count])
			fmt.Println("Reply from server: runtime=", content)
			fmt.Println("RTT = ", rtt, "ms")
		case 5:
			pconn.Close()
			fmt.Println("Bye bye~")
			os.Exit(0)
		default:
			fmt.Println("Please enter number between 1 and 5")
		}
	}
}

func openSignalChannel(ment string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		sig := <-sigs
		if sig.String() == "interrupt" {
			fmt.Println()
			fmt.Println(ment)
		} else {
			fmt.Println("unhandled signal")
		}
		close(sigs)
		os.Exit(0)
	}()
}

func createJsonMessage(opt int, content string) []byte {
	req_msg := message{Option: opt, Content: content}
	// encode struct to json byte array
	json_msg, err := json.Marshal(req_msg)
	if err != nil {
		log.Fatal("json encode error: ", err)
	}
	return json_msg
}

/**
*
 */
func decodeJsonMessage(buffer []byte) (int, string) {
	var res_msg message
	json.Unmarshal(buffer, &res_msg)

	return res_msg.Option, res_msg.Content
}