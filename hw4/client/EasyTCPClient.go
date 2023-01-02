/**
 * TCPClient.go
 * Name: Jeong Yong Jun
 * StudentID: 20173875
 **/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

/**
* Struct RequestMessage
* Format of Requestmessage
* Command: a value that indicates what to do. Between 0 and 7.
* SenderName: sender nickname
* ReceiverName: receiver nickname
* Content: Requestmessage content
 */
type RequestMessage struct {
	Command      byte   `json:"command"`
	SenderName   string `json:"senderName"`
	ReceiverName string `json:"receiverName"`
	Content      string `json:"content"`
}

/**
* Struct ResponseMessage
* Format of Responsemessage
* Option: a value that indicates what to do. Between 0 and 7. Error for 10, 11
* Content: Requestmessage content
 */
type ResponseMessage struct {
	Option  int    `json:"option"`
	Content string `json:"content"`
}

func main() {

	serverName := "localhost"
	serverPort := "33875"

	nickname := os.Args[1]
	conn, err := net.Dial("tcp", serverName+":"+serverPort)

	openSignalChannel(conn, nickname, "gg~")
	if err != nil {
		log.Fatalln("tcp connection error: ", err)
	}

	defer conn.Close()

	readChannel := make(chan ResponseMessage)
	go startMessageReader(conn, readChannel)
	conn.Write(createRequestMessage(0, nickname, " ", " "))
	firstMessage := <-readChannel

	if firstMessage.Option == 10 || firstMessage.Option == 11 {
		fmt.Println(firstMessage.Content)
		fmt.Println("gg~")
		conn.Close()
		os.Exit(0)
	} else {
		fmt.Println(firstMessage.Content)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// choose input option
		scanner.Scan()
		input := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Fatalln("Scanning input error: ", err)
		}
		if input == "" {
			fmt.Println("Write some text for message")
			continue
		}
		var requestMessage []byte
		var responseMessage ResponseMessage

		// process command
		if input[0] == '\\' {
			if input == "\\list" { // list
				requestMessage = createRequestMessage(2, nickname, nickname, " ")
				conn.Write(requestMessage)

				responseMessage = <-readChannel
				fmt.Println(responseMessage.Content)
			} else if strings.HasPrefix(input, "\\dm") { // dm
				if strings.Count(input, " ") < 2 {
					fmt.Println("invalid command. dm format is incorrect.")
					continue
				}
				dmInput := input[strings.Index(input, " ")+1:]
				receiverName := dmInput[:strings.Index(dmInput, " ")]
				dmMessage := dmInput[strings.Index(dmInput, " ")+1:]
				if dmMessage == "" {
					fmt.Println("Write some text for direct message")
					continue
				}
				requestMessage = createRequestMessage(3, nickname, receiverName, dmMessage)
				conn.Write(requestMessage)

			} else if input == "\\exit" { // exit
				requestMessage = createRequestMessage(4, nickname, " ", " ")
				conn.Write(requestMessage)
				conn.Close()
				fmt.Println("gg~")
				os.Exit(0)
			} else if input == "\\ver" { // ver
				requestMessage = createRequestMessage(5, nickname, nickname, " ")
				conn.Write(requestMessage)

				responseMessage = <-readChannel
				fmt.Println(responseMessage.Content)
			} else if input == "\\rtt" { // rtt
				requestMessage = createRequestMessage(6, nickname, nickname, " ")
				requestTime := time.Now()
				conn.Write(requestMessage)

				responseMessage = <-readChannel
				rtt := float64(time.Since(requestTime)) / float64(time.Millisecond)
				fmt.Println("RTT = ", rtt, "ms")
			} else {
				fmt.Println("invalid command")
				continue
			}
		} else { // broadcast message
			requestMessage = createRequestMessage(1, nickname, " ", input)
			conn.Write(requestMessage)
		}
	}
}

/**
* oepnSignalChannel
* conn: tcp connection
* nickname: nickname
* ment: good bye ment
* When signal Ctrl+C inserted, send exit message, print ment and exit process
 */
func openSignalChannel(conn net.Conn, nickname string, ment string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		sig := <-sigs
		if sig.String() == "interrupt" {
			requestMessage := createRequestMessage(4, nickname, " ", " ")
			conn.Write(requestMessage)
			conn.Close()
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
func createRequestMessage(command byte, senderName string, receiverName string, content string) []byte {
	msg := RequestMessage{Command: command, SenderName: senderName, ReceiverName: receiverName, Content: content}
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("json encode error: ", err)
	}
	return jsonMessage
}

/**
* decodeJsonMessage
* buf: encoded json byte array
* return: json decoded Message struct
* get json byte array and decode it to Message struct
 */
func decodeJsonMessage(buffer []byte) ResponseMessage {
	var msg ResponseMessage
	json.Unmarshal(buffer, &msg)

	return msg
}

/**
* startMessageReader
* conn: network connection
* readChannel: channel to send buffer data
* Read buffer, decode buffer data, send message through channel
 */
func startMessageReader(conn net.Conn, readChannel chan ResponseMessage) {
	for {
		buffer := make([]byte, 4096)
		count, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("server closes connection", err)
				fmt.Println("gg~")
				conn.Close()
				os.Exit(0)
			}
		}
		jsonMessage := decodeJsonMessage(buffer[:count])
		switch jsonMessage.Option {
		case 0, 2, 5, 6, 10, 11:
			readChannel <- jsonMessage
		default:
			fmt.Println(jsonMessage.Content)
		}
	}
}
