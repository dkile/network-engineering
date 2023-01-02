/**
 * UDPServer.go
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

/**
* Struct Message
* Format of message
* Option: a value that indicates what to do. Between 0 and 4
* Conatent: message content
 */
type Message struct {
    Option int `json:"option"`
    Content string `json:"content"`
}

func main() {
    
    openSignalChannel("Bye bye~")

    startTime := time.Now()
    serverPort := "23875"

    pconn, _:= net.ListenPacket("udp", ":"+serverPort)
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)

    buffer := make([]byte, 4096)

    requestCount := 0

    for {
        count, remoteAddr, err:= pconn.ReadFrom(buffer)
        if err != nil {
            log.Fatal("buffer reading error: ", err)
        }
        requestCount += 1

        requestMessage := decodeJsonMessage(buffer[:count])
        
        fmt.Printf("UDP message from %s\n", remoteAddr.String())

        switch	requestMessage.Option {
        case 1:
            pconn.WriteTo(createResponseMessage(1, strings.ToUpper(requestMessage.Content)), remoteAddr)
        case 2:
            pconn.WriteTo(createResponseMessage(2, remoteAddr.String()), remoteAddr)
        case 3:
            pconn.WriteTo(createResponseMessage(3, strconv.Itoa(requestCount)), remoteAddr)
        case 4:
            pconn.WriteTo(createResponseMessage(4, formatDuration(time.Now().Sub(startTime))), remoteAddr)
        default:
            pconn.WriteTo(createResponseMessage(0, ""), remoteAddr)
        }
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
		log.Println("ResponseMessage encoding error", err)
	}
	return jsonMessage
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