/**
 * TCPServer.go
 * Name: Jeong Yong Jun
 * StudentID: 20173875
 **/

package main

import (
	"context"
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

	startTime := time.Now()
	serverPort := "33875"
	requestCount := 0
	connCount := 0
	connMap := make(map[string]int)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context, interval time.Duration) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Println("Number of connected clients =", len(connMap))
			}
		}
	}(ctx, time.Second * 60)

	openSignalChannel("Bye bye~")

	listener, _ := net.Listen("tcp", ":"+serverPort)
	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("connection error: ", err)
			continue
		}
		defer conn.Close()

		go func() {
			
			remoteAddr := conn.RemoteAddr().String()
			connCount += 1
			connMap[remoteAddr] = connCount
			fmt.Println(remoteAddr)
			fmt.Println("Client ", connCount, " connected. Number of connected clients = ", len(connMap))

			buffer := make([]byte, 4096)
			for {
				count, err := conn.Read(buffer)
				if err != nil {
					if err == io.EOF {
						// log.Println("Client closes connection.")
						id := connMap[remoteAddr]
						delete(connMap, remoteAddr)
						fmt.Println("Client", id, "disconnected. Number of connected clients = ", len(connMap))
						conn.Close()
					} else {
						log.Println("buffer reading error: ", err)
					}
					return
				}
																															
				requestCount += 1

				requestMessage := decodeJsonMessage(buffer[:count])

				switch requestMessage.Option {
				case 1:
					conn.Write(createResponseMessage(1, strings.ToUpper(requestMessage.Content)))
				case 2:
					conn.Write(createResponseMessage(2, conn.RemoteAddr().String()))
				case 3:
					conn.Write(createResponseMessage(3, strconv.Itoa(requestCount)))
				case 4:
					conn.Write(createResponseMessage(4, formatDuration(time.Now().Sub(startTime))))
				default:
					conn.Write(createResponseMessage(0, ""))
				}
			}
		}()
	}
}

/**
* createResponseMessage
* option: response message option
* content: response message content
* return: json encoded byte array
* create Message struct with option, content and encode it by json
*/
func createResponseMessage(option int, content string) []byte {
	msg := Message{Option: option, Content: content}
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
		msg.Option = 0
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

	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

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
* formatDuration
* d: duration to format
* format nano second duration to "HH:MM:SS"
*/
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
