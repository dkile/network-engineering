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

/**
* Struct ReadChanMessage
* Format of ReadChanMessage
* Broadcast: broadcast(true) or direct(false)
* Nickname: if broadcast is true, send message except nickname. if not, to nickname
* Message: message content
 */
type ReadChanMessage struct {
	Broadcast bool
	Nickname  string
	Message   []byte
}

func main() {
	serverPort := "33875"
	userList := make(map[string]string)

	openSignalChannel("gg~")

	listener, _ := net.Listen("tcp", ":"+serverPort)
	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	defer listener.Close()
	readChan := make(chan ReadChanMessage)

	for {
		conn, err := listener.Accept()
		serverIP := conn.LocalAddr().(*net.TCPAddr).IP.String()
		if err != nil {
			log.Println("connection error: ", err)
			continue
		}

		go func() {
			defer conn.Close()
			runChan := make(chan bool)
			go connReader(conn, runChan, readChan, userList)
			remoteAddr := conn.RemoteAddr().String()

			buffer := make([]byte, 4096)
			count, err := conn.Read(buffer)
			if err != nil {
				log.Println("buffer reading error: ", err)
				return
			}
			firstRequestMessage, _ := decodeJsonMessage(buffer[:count])
			nickname := firstRequestMessage.SenderName
			_, exists := userList[nickname]

			if len(userList) >= 8 { // chat room full
				conn.Write(createResponseMessage(10, "chatting room full. cannot connect"))
				time.Sleep(time.Millisecond * 5)
				return
			}
			if exists { // duplicate nickname
				conn.Write(createResponseMessage(11, "that nickname is already used by another user. cannot connect."))
				time.Sleep(time.Millisecond * 5)
				return
			}

			// register nickname
			userList[nickname] = remoteAddr
			conn.Write(createResponseMessage(0, "Welcome "+nickname+" to CAU network class chat room at <"+serverIP+":"+serverPort+">. There are <"+strconv.Itoa(len(userList))+"> users connected."))
			fmt.Println(nickname + " joined from <" + remoteAddr + ">. There are " + strconv.Itoa(len(userList)) + " users connected.")

			runChan <- true // start connReader

			// send response message by broadcast or direct
			for {
				select {
				case readChanMessage := <-readChan:
					if readChanMessage.Broadcast {
						if nickname != readChanMessage.Nickname {
							conn.Write(readChanMessage.Message)
						}
					} else {
						if nickname == readChanMessage.Nickname {
							conn.Write(readChanMessage.Message)
						}
					}
				case <-runChan:
					time.Sleep(time.Millisecond * 3)
					delete(userList, nickname)
					readChanMessage := makeReadChanMessage(true, nickname, createResponseMessage(4, nickname+" left. There are "+strconv.Itoa(len(userList))+" users now."))
					fmt.Println(nickname + " left. There are " + strconv.Itoa(len(userList)) + " users now.")
					for i := 0; i < len(userList); i++ {
						readChan <- readChanMessage
					}
					time.Sleep(time.Millisecond * 3)
					conn.Close()
					return
				}
			}
		}()
	}
}

/**
* createResponseMessage
* option: response Responsemessage option
* content: response Responsemessage content
* return: json encoded byte array
* create ResponseMessage struct with command, content and encode it by json
 */
func createResponseMessage(option int, content string) []byte {
	msg := ResponseMessage{Option: option, Content: content}
	jsonResponseMessage, err := json.Marshal(msg)
	if err != nil {
		log.Println("Response message encoding error", err)
	}
	return jsonResponseMessage
}

/**
* decodeJsonMessage
* buf: encoded json byte array
* return: json decoded RequestMessage struct
* get json byte array and decode it to RequestMessage struct
 */
func decodeJsonMessage(buf []byte) (RequestMessage, error) {
	var msg RequestMessage
	err := json.Unmarshal(buf, &msg)
	if err != nil {
		log.Println("Request message decoding error", err)
	}
	return msg, err
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

/**
*	connReader
* conn: tcp connection
* runChannel: on/off channel for connReader
* readChannel: message data transfer channel
* userList: list of users
* Read buffer and do specific actions(sending response message through read channel) by command
 */
func connReader(conn net.Conn, runChannel chan bool, readChannel chan ReadChanMessage, userList map[string]string) {
	defer conn.Close()
	buffer := make([]byte, 4096)
	<-runChannel
	for {
		count, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("client is dead")
				runChannel <- false
				break
			}
		}
		requestMessage, err := decodeJsonMessage(buffer[:count])
		if err != nil {
			continue
		}
		senderName := requestMessage.SenderName

		var readChanMessage ReadChanMessage

		switch requestMessage.Command {
		case 1:
			readChanMessage = makeReadChanMessage(true, senderName, createResponseMessage(1, senderName+"> "+requestMessage.Content))
		case 2:
			responseString := ""
			for nick, ra := range userList {
				ip := strings.Split(ra, ":")[0]
				port := strings.Split(ra, ":")[1]
				responseString += ("<" + nick + " " + ip + ", " + port + "> ")
			}
			readChanMessage = makeReadChanMessage(false, requestMessage.ReceiverName, createResponseMessage(2, "User list: "+responseString))
		case 3:
			responseString := "from: " + senderName + "> " + requestMessage.Content
			_, exists := userList[requestMessage.ReceiverName]
			if exists {
				readChanMessage = makeReadChanMessage(false, requestMessage.ReceiverName, createResponseMessage(3, responseString))
			} else {
				readChanMessage = makeReadChanMessage(false, senderName, createResponseMessage(12, "No such user"))
			}
		case 4:
			runChannel <- false
			return
		case 5:
			readChanMessage = makeReadChanMessage(false, requestMessage.ReceiverName, createResponseMessage(5, "server version is 1.0"))
		case 6:
			readChanMessage = makeReadChanMessage(false, requestMessage.ReceiverName, createResponseMessage(6, ""))
		default:
			fmt.Println("invalid command")
		}
		for i := 0; i < len(userList); i++ {
			readChannel <- readChanMessage
		}

		// if message contains "i hate professor", disconnect the user
		if strings.Contains(strings.ToLower(requestMessage.Content), "i hate professor") {
			runChannel <- false
			return
		}
	}
}

/**
* makeReadChanMessage
* return new ReadChanMessage
 */
func makeReadChanMessage(broadcast bool, nickname string, message []byte) ReadChanMessage {
	return ReadChanMessage{
		Broadcast: broadcast,
		Nickname:  nickname,
		Message:   message,
	}
}
