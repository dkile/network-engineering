/**
 * TCPServer.go
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

type message struct {
	Option  int    `json:"option"`
	Content string `json:"content"`
}

func main() {

	openSignalChannel("Bye bye~")

	start_time := time.Now()
	serverPort := "33875"

	listener, _ := net.Listen("tcp", ":"+serverPort)
	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	defer listener.Close()

	req_cnt := 0

	for {
		conn, err := listener.Accept()
		time_ch := make(chan time.Time)
		if err != nil {
			log.Println("connection error: ", err)
		}

		go func() {
			buffer := make([]byte, 4096)
			for {
				count, err := conn.Read(buffer)
				if err != nil {
					if err == io.EOF {
							log.Println("connection closed from client ", conn.RemoteAddr().String())
						} else {
							log.Println("Read error: ", err)
						}
						time_ch <- time.Time{}
						return
					}
				time_ch <- time.Now()
				fmt.Println("connect with ", conn.RemoteAddr().String())

				req_cnt += 1

				var req_msg message
				json.Unmarshal(buffer[:count], &req_msg)

				switch req_msg.Option {
				case 1:
					upper_sentence := strings.ToUpper(req_msg.Content)
					res_msg := message{Option: 1, Content: upper_sentence}
					json_msg, _ := json.Marshal(res_msg)
					conn.Write(json_msg)
					fmt.Println("case 1")
				case 2:
					res_msg := message{Option: 2, Content: conn.RemoteAddr().String()}
					json_msg, _ := json.Marshal(res_msg)
					conn.Write(json_msg)
					fmt.Println("case 2")
				case 3:
					res_msg := message{Option: 3, Content: strconv.Itoa(req_cnt)}
					json_msg, _ := json.Marshal(res_msg)
					conn.Write(json_msg)
					fmt.Println("case 3")
				case 4:
					duration := time.Now().Sub(start_time)
					res_msg := message{Option: 4, Content: fmtDuration(duration)}
					json_msg, _ := json.Marshal(res_msg)
					conn.Write(json_msg)
					fmt.Println("case 4")
				default:
					conn.Write([]byte("error"))
					fmt.Println("case 5")
				}
			}
		}()

		go func() {
			timer := time.NewTimer(time.Second * 30)
			fmt.Println(conn, "timer started")
			for {
				select {
				case req_time:= <- time_ch:
					if req_time.IsZero() {
						// Use SetLinger to force close the connection
						err := conn.(*net.TCPConn).SetLinger(0)
						if err != nil {
							log.Printf("Error when setting linger: %s", err)
						}
						fmt.Println("connection ", conn, " close")
						conn.Close()
						return
					} else {
						timer.Reset(time.Second * 30)
						fmt.Println(conn, "timer reset")
					}
				case <- timer.C:
					fmt.Println(conn, "timer end")
					fmt.Println("conenction", conn, "close")
					conn.Close()
					return
				}
		}
		}()
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
			close(sigs)
			os.Exit(0)
		} else {
			fmt.Println("unhandled signal")
		}
	}()
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
