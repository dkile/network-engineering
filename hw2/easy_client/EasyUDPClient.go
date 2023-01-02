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

type Message struct {
	Option  int    `json:"option"`
	Content string `json:"content"`
}

func main() {

	openSignalChannel("Bye bye~")

	// define server name, port
	serverName := "localhost"
	serverPort := "23875"

	serverAddr, _ := net.ResolveUDPAddr("udp", serverName+":"+serverPort)

	// create UDP socket
	pconn, _ := net.ListenPacket("udp", ":")

	localAddr := pconn.LocalAddr().(*net.UDPAddr)
	fmt.Println("Client is running on port", localAddr.Port)

	defer pconn.Close()

	for {

		// choose input option
		var inputOption string
		fmt.Println("<Menu>")
		fmt.Println("1) convert text to UPPER-case")
		fmt.Println("2) get my IP address and port number")
		fmt.Println("3) get server request count")
		fmt.Println("4) get server running time")
		fmt.Println("5) exit")
		fmt.Printf("Input option: ")
		_, err := fmt.Scanln(&inputOption)
		if err != nil {
			log.Fatalln("Scanning input error: ", err)
		}
		option, _ := strconv.Atoi(inputOption)

		var requestMessage []byte

		switch option {
		case 1:
			var inputSentence string
			fmt.Printf("Input lowercase sentence: ")
			fmt.Scanln(&inputSentence)
			requestMessage = createRequestMessage(1, inputSentence)
		case 2, 3, 4:
			requestMessage = createRequestMessage(option, "")
		case 5:
			pconn.Close()
			fmt.Println("Bye bye~")
			os.Exit(0)
		default:
			fmt.Println("Please enter number between 1 and 5")
			continue
		}

		req_time := time.Now()

		pconn.WriteTo(requestMessage, serverAddr)

		buffer := make([]byte, 4096)
		pconn.SetReadDeadline(time.Now().Add(time.Second * 5))
		count, _, err := pconn.ReadFrom(buffer)

		rtt := float64(time.Now().Sub(req_time)) / float64(time.Millisecond)

		if err != nil {
			log.Println("No response from server")
			continue
		}

		responseMessage := decodeJsonMessage(buffer[:count])
		fmt.Println(getResult(responseMessage))
		fmt.Println("RTT = ", rtt, "ms")
	}
}

/**
* oepnSignalChannel
* ment: good bye ment
* When signal Ctrl+C inserted, print ment and exit process
*/
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

/**
* createRequestMessage
* option: response message option
* content: response message content
* return: json encoded byte array
* create Message struct with option, content and encode it by json
*/
func createRequestMessage(opt int, content string) []byte {
	msg := Message{Option: opt, Content: content}
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("json encode error: ", err)
	}
	return jsonMessage
}

/**
* getResult
* reply: Message struct
* return: string to print
* classify message by option and make string to print
*/
func getResult(reply Message) string {
	result := "Reply from server: "
	switch reply.Option {
	case 1, 3, 4:
		result += reply.Content
	case 2:
		ip := strings.Split(reply.Content, ":")[0]
		port := strings.Split(reply.Content, ":")[1]
		result += ("client IP=" + ip + " port=" + port)
	default:
		result += "error occured"
	}
	return result
}

/**
* decodeJsonMessage
* buf: encoded json byte array
* return: json decoded Message struct
* get json byte array and decode it to Message struct
*/
func decodeJsonMessage(buffer []byte) Message {
	var msg Message
	json.Unmarshal(buffer, &msg)

	return msg
}