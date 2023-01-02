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
	"strconv"
	"strings"
	"syscall"
	"time"
)

/**
* Struct Message
* Format of message
* Option: a value that indicates what to do. Between 0 and 4
* Conatent: message content
 */
type Message struct {
	Option  int    `json:"option"`
	Content string `json:"content"`
}

func main() {
	openSignalChannel("Bye bye~")

	serverName := "localhost"
	serverPort := "33875"

	conn, err := net.Dial("tcp", serverName+":"+serverPort)
	if err != nil {
		log.Fatalln("tcp connection error: ", err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	fmt.Printf("Client is running on port %d\n", localAddr.Port)

	readChannel := make(chan []byte)
	go startMessageReader(conn, readChannel)

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
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Printf("Input lowercase sentence: ")
			if scanner.Scan() {
				inputSentence := scanner.Text()
				requestMessage = createRequestMessage(1, inputSentence)
			}
		case 2, 3, 4:
			requestMessage = createRequestMessage(option, "")
		case 5:
			conn.Close()
			fmt.Println("Bye bye~")
			os.Exit(0)
		default:
			fmt.Println("Please enter number between 1 and 5")
			continue
		}

		requestTime := time.Now()

		conn.Write(requestMessage)
		jsonMessage := <-readChannel

		rtt := float64(time.Now().Sub(requestTime)) / float64(time.Millisecond)

		fmt.Println(getResult(decodeJsonMessage(jsonMessage)))
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

/**
* startMessageReader
* conn: network connection
* readChannel: channel to send buffer data
* Read buffer and send buffer data by channel
 */
func startMessageReader(conn net.Conn, readChannel chan []byte) {
	for {
		buffer := make([]byte, 4096)
		count, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Fatalln("server closes connection")
			}
		}
		readChannel <- buffer[:count]
	}
}
