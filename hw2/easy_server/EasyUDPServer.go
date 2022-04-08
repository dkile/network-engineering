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

type message struct {
    Option int `json:"option"`
    Content string `json:"content"`
}

func main() {
    
    openSignalChannel("Bye bye~")

    start_time := time.Now()
    serverPort := "23875"

    pconn, _:= net.ListenPacket("udp", ":"+serverPort)
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)

    buffer := make([]byte, 1024)

    var req_msg message

    req_cnt := 0

    for {
        count, r_addr, err:= pconn.ReadFrom(buffer)
        if err != nil {
            log.Fatal("Read error: ", err)
        }
        req_cnt += 1
        json.Unmarshal(buffer[:count], &req_msg)

        
        fmt.Printf("UDP message from %s\n", r_addr.String())
        // fmt.Printf("res msg option: %d\n", req_msg.Option)
        // fmt.Printf("res msg contetnt: %s\n", req_msg.Content)

        switch req_msg.Option {
        case 1:
            upper_sentence := strings.ToUpper(req_msg.Content)
            res_msg := &message{Option: 1, Content: upper_sentence}
            json_msg, _ := json.Marshal(res_msg)
            pconn.WriteTo(json_msg, r_addr)
        case 2:
            res_msg := &message{Option: 2, Content: r_addr.String()}
            json_msg, _ := json.Marshal(res_msg)
            pconn.WriteTo(json_msg, r_addr)
        case 3:
            res_msg := &message{Option: 3, Content: strconv.Itoa(req_cnt)}
            json_msg, _ := json.Marshal(res_msg)
            pconn.WriteTo(json_msg, r_addr)
        case 4:
            duration := time.Now().Sub(start_time)
            res_msg := &message{Option: 4, Content: fmtDuration(duration)}
            json_msg, _ := json.Marshal(res_msg)
            pconn.WriteTo(json_msg, r_addr)
        default:
            pconn.WriteTo([]byte("error"), r_addr)
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

func fmtDuration(d time.Duration) string {
    d = d.Round(time.Second)
    h := d / time.Hour
    d -= h * time.Hour
    m := d / time.Minute
    d -= m * time.Minute
    s := d / time.Second
    return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}