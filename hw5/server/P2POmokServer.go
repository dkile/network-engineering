/**
 * TCPServer.go
 * Name: Jeong Yong Jun
 * StudentID: 20173875
 **/

package main

import (
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
 * Content: chat message or who's first
 * Info: client information
 */
type Message struct {
	Content string     `json:"content"`
	Info    ClientInfo `json:"info"`
}

/**
 * Struct ClientInfo
 * Format of client information
 * Nickname: user nickname
 * IP: user ip
 * Port: UDP Port
 */
type ClientInfo struct {
	Nickname string `json:"nickname"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
}

func main() {

	serverPort := "53875"

	openSignalChannel("Bye~")

	listener, _ := net.Listen("tcp", ":"+serverPort)
	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	enterSignal := make(chan ClientInfo)
	outputChannel := make(chan []Message)

	go matchUser(enterSignal, outputChannel, 2)
	defer listener.Close()
	defer close(enterSignal)
	defer close(outputChannel)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("connection error: ", err)
			continue
		}
		go processClient(conn, enterSignal, outputChannel)
	}
}

/**
 * processClient
 * conn: tcp connection
 * enterSignal: channel to notify entered user
 * outputChannel: channel for matched users
 * process client message
 */
func processClient(conn net.Conn, enterSignal chan<- ClientInfo, outputChannel <-chan []Message) {
	buffer := make([]byte, 4096)
	count, err := conn.Read(buffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Client closes connection.")
			conn.Close()
		} else {
			log.Println("buffer reading error: ", err)
		}
		return
	}
	remoteAddr := conn.RemoteAddr().String()
	myInfo := decodeJsonMessage(buffer[:count]).Info
	myInfo.IP = strings.Split(remoteAddr, ":")[0]
	fmt.Println(myInfo.Nickname + " joined from " + remoteAddr + ". UDP port " + myInfo.Port)
	enterSignal <- myInfo
	conn.Write(createResponseMessage(Message{Content: "welcome " + myInfo.Nickname + " to p2p-omok server at " + remoteAddr, Info: ClientInfo{}}))
	userList := <-outputChannel
	remoteInfos := userFilter(userList, func(v Message) bool {
		return v.Info.Nickname != myInfo.Nickname
	})
	time.Sleep(time.Millisecond * 5)
	conn.Write(createResponseMessage(remoteInfos[0]))
	conn.Close()
}

func userFilter(users []Message, f func(Message) bool) []Message {
	result := make([]Message, 0)
	for _, user := range users {
		if f(user) {
			result = append(result, user)
		}
	}
	return result
}

/**
 * matchUser
 * enterSignal: channel to notify entered user
 * outputChannel: channel for matched users
 * numUser: max user number
 * match users by numUser
 */
func matchUser(enterSignal <-chan ClientInfo, outputChannel chan<- []Message, numUser int) {
	userList := make([]Message, 0, numUser)
	userNum := 0
	for {
		user := <-enterSignal
		userList = append(userList, Message{Content: strconv.Itoa(userNum), Info: user})
		userNum++
		if len(userList) < numUser {
			fmt.Println("1 user connected. waiting for other client...")
		}
		if len(userList) == numUser {
			top := len(userList) - 1
			user1 := userList[top].Info
			user2 := userList[top-1].Info
			for i := 0; i < numUser; i++ {
				outputChannel <- userList
			}
			userList = userList[:0]
			fmt.Println("2 user connected, notifying", user1.Nickname, "and", user2.Nickname)
			fmt.Println(user1.Nickname, "and", user2.Nickname, "disconnected")
			userNum = 0
		}
	}
}

/**
 * createResponseMessage
 * message: message to send
 * return: json encoded byte array
 * create Message struct with option, content and encode it by json
 */
func createResponseMessage(message Message) []byte {
	msg := message
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Println("Response message encoding error", err)
	}
	return jsonMessage
}

/**
 * decodeJsonMessage
 * buf: encoded json byte array
 * return: json decoded Message struct
 * get json byte array and decode it to Message struct
 */
func decodeJsonMessage(buf []byte) Message {
	var msg Message
	err := json.Unmarshal(buf, &msg)
	if err != nil {
		log.Println("Request message decoding error", err)
	}

	return msg
}

/**
 * oepnSignalChannel
 * ment: good bye ment
 * When signal Ctrl+C inserted, print ment and exit process
 */
func openSignalChannel(ment string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		sig := <-sigs
		if sig.String() == "interrupt" {
			fmt.Println()
			fmt.Println(ment)
			close(sigs)
			os.Exit(0)
		} else {
			log.Println("unhandled signal")
		}
	}()
}
