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
	"os/exec"
	"os/signal"
	"runtime"
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

/**
 * Struct P2PMessage
 * Format of p2p message
 * Command: user command - 0: chat, 1: put stone, 2: gg, 3: exit
 * Content: chat message
 */
type P2PMessage struct {
	Command byte   `json:"command"`
	Content string `json:"content"`
}

type Point struct {
	x    int
	y    int
	turn int
}

const (
	Row = 10
	Col = 10
)

var gameOver bool

type Board [][]int

func main() {
	nickname := os.Args[1]
	udpConn, _ := net.ListenPacket("udp", ":")
	localAddr := udpConn.LocalAddr().(*net.UDPAddr)
	defer udpConn.Close()

	oponentInfo := findOponent(createClientInfo("", localAddr.Port, nickname))
	myTurn := 0
	gameOver = false
	fmt.Println("Found oponent " + oponentInfo.Info.IP + ":" + oponentInfo.Info.Port)
	if oponentInfo.Content == "0" {
		fmt.Println("You play second")
		myTurn = 1
	} else if oponentInfo.Content == "1" {
		fmt.Println("You play first")
		myTurn = 0
	}

	oponentAddr, _ := net.ResolveUDPAddr("udp", oponentInfo.Info.IP+":"+oponentInfo.Info.Port)
	p2pChannel := make(chan P2PMessage)
	boardChannel := make(chan Point)
	timeChannel := make(chan time.Time)
	openSignalChannel(udpConn, oponentAddr, "Bye~")
	go startP2PReader(udpConn, p2pChannel)
	go processP2PMessage(p2pChannel, boardChannel, oponentInfo)
	go manageBoard(boardChannel, timeChannel, myTurn)
	go runTimer(udpConn, oponentAddr, timeChannel, time.Second*10)

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
		// process command
		if input[0] == '\\' {
			if input[:2] == "\\\\" {
				if gameOver {
					continue
				}
				x, _ := strconv.Atoi(strings.Split(input[2:], " ")[0])
				y, _ := strconv.Atoi(strings.Split(input[2:], " ")[1])
				boardChannel <- Point{x: x, y: y, turn: myTurn}

				requestMessage = createP2PMessage(1, input[2:])
				udpConn.WriteTo(requestMessage, oponentAddr)
				timeChannel <- time.Time{}
			} else if input == "\\gg" {
				if gameOver {
					continue
				}
				requestMessage = createP2PMessage(2, "")
				udpConn.WriteTo(requestMessage, oponentAddr)
				fmt.Println("You lose")
				gameOver = true
			} else if input == "\\exit" {
				requestMessage = createP2PMessage(3, "")
				udpConn.WriteTo(requestMessage, oponentAddr)
				if !gameOver {
					fmt.Println("You lose")
				}
				fmt.Println("Bye~")
				os.Exit(0)
			} else {
				fmt.Println("invalid command")
				continue
			}
		} else {
			requestMessage = createP2PMessage(0, input)
			udpConn.WriteTo(requestMessage, oponentAddr)
		}
	}
}

/**
 * startP2PReader
 * conn: udp connection
 * p2pChannel: channel for p2p message
 * Read buffer and send buffer data through channel
 */
func startP2PReader(conn net.PacketConn, p2pChannel chan<- P2PMessage) {
	for {
		buffer := make([]byte, 4096)
		count, _, err := conn.ReadFrom(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("You win")
				conn.Close()
			}
		}
		p2pMessage := decodeJsonToP2P(buffer[:count])
		p2pChannel <- p2pMessage
	}
}

/**
 * processP2PMessage
 * p2pChannel: channel for p2p message struct
 * boardChannel: channel for updating board
 * oponentInfo: oponent information
 * process p2p message by command
 */
func processP2PMessage(p2pChannel <-chan P2PMessage, boardChannel chan<- Point, oponentInfo Message) {
	for {
		p2pMessage := <-p2pChannel
		switch p2pMessage.Command {
		case 0:
			fmt.Println(oponentInfo.Info.Nickname + "> " + p2pMessage.Content)
		case 1:
			if gameOver {
				continue
			}
			turn, _ := strconv.Atoi(oponentInfo.Content)
			x, _ := strconv.Atoi(strings.Split(p2pMessage.Content, " ")[0])
			y, _ := strconv.Atoi(strings.Split(p2pMessage.Content, " ")[1])
			boardChannel <- Point{x: x, y: y, turn: turn}
		case 2:
			if gameOver {
				continue
			}
			fmt.Println("You win")
			gameOver = true
		case 3:
			fmt.Println("Oponent exits.")
			if gameOver {
				continue
			}
			fmt.Println("You win")
			gameOver = true
		}
	}
}

func createP2PMessage(command byte, content string) []byte {
	msg := P2PMessage{Command: command, Content: content}
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("json encode error: ", err)
	}
	return jsonMessage
}

func initializeBoard() Board {
	board := Board{}
	for i := 0; i < Row; i++ {
		var tempRow []int
		for j := 0; j < Col; j++ {
			tempRow = append(tempRow, 0)
		}
		board = append(board, tempRow)
	}
	return board
}

/**
 * manageBoard
 * boardChannel: channel for Point
 * timeChannel: timer channel
 * myTurn: client turn number
 * initialize and update omok board
 */
func manageBoard(boardChannel chan Point, timeChannel chan time.Time, myTurn int) {
	board := initializeBoard()
	printBoard(board)
	count, win, turn := 0, 0, 0
	for {
		point := <-boardChannel
		if point.x < 0 || point.y < 0 || point.x >= Row || point.y >= Col {
			if point.turn == myTurn {
				fmt.Println("error, out of bound!")
			}
			time.Sleep(1 * time.Second)
			continue
		} else if board[point.x][point.y] != 0 {
			if point.turn == myTurn {
				fmt.Println("error, already used!")
			}
			time.Sleep(1 * time.Second)
			continue
		} else if point.turn != turn {
			time.Sleep(1 * time.Second)
			continue
		}

		if point.turn == 0 {
			board[point.x][point.y] = 1
		} else {
			board[point.x][point.y] = 2
		}

		clear()
		printBoard(board)

		win = checkWin(board, point.x, point.y)
		if win != -1 {
			if win == myTurn+1 {
				fmt.Println("You win")
			} else {
				fmt.Println("You lose")
			}
			gameOver = true
			break
		}

		count += 1
		if count == Row*Col {
			fmt.Printf("draw!\n")
			break
		}
		turn = (turn + 1) % 2
		if point.turn != myTurn {
			timeChannel <- time.Now()
		}
	}
}

func printBoard(b Board) {
	fmt.Print("   ")
	for j := 0; j < Col; j++ {
		fmt.Printf("%2d", j)
	}

	fmt.Println()
	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()

	for i := 0; i < Row; i++ {
		fmt.Printf("%d |", i)
		for j := 0; j < Col; j++ {
			c := b[i][j]
			if c == 0 {
				fmt.Print(" +")
			} else if c == 1 {
				fmt.Print(" 0")
			} else if c == 2 {
				fmt.Print(" @")
			} else {
				fmt.Print(" |")
			}
		}

		fmt.Println(" |")
	}

	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()
}

func checkWin(b Board, x, y int) int {
	lastStone := b[x][y]
	startX, startY, endX, endY := x, y, x, y

	// Check X
	for startX-1 >= 0 && b[startX-1][y] == lastStone {
		startX--
	}
	for endX+1 < Row && b[endX+1][y] == lastStone {
		endX++
	}

	// Check Y
	startX, startY, endX, endY = x, y, x, y
	for startY-1 >= 0 && b[x][startY-1] == lastStone {
		startY--
	}
	for endY+1 < Row && b[x][endY+1] == lastStone {
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 1
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && startY-1 >= 0 && b[startX-1][startY-1] == lastStone {
		startX--
		startY--
	}
	for endX+1 < Row && endY+1 < Col && b[endX+1][endY+1] == lastStone {
		endX++
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 2
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && endY+1 < Col && b[startX-1][endY+1] == lastStone {
		startX--
		endY++
	}
	for endX+1 < Row && startY-1 >= 0 && b[endX+1][startY-1] == lastStone {
		endX++
		startY--
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	return -1
}

func clear() {
	fmt.Printf("%s", runtime.GOOS)

	clearMap := make(map[string]func()) //Initialize it
	clearMap["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearMap["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	value, ok := clearMap[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                             //if we defined a clearMap func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clearMap terminal screen :(")
	}
}

func createClientInfo(ip string, port int, nickname string) ClientInfo {
	return ClientInfo{Nickname: nickname, IP: ip, Port: strconv.Itoa(port)}
}

/**
 * findOponent
 * myInfo: my udp information to send to server
 * return: oponent information in Message struct
 * send my information to server and find oponent
 */
func findOponent(myInfo ClientInfo) Message {
	var result Message
	serverName := "localhost"
	serverPort := "53875"

	conn, err := net.Dial("tcp", serverName+":"+serverPort)
	if err != nil {
		log.Fatalln("tcp connection error: ", err)
	}
	conn.Write(createRequestMessage("", myInfo))
	readChannel := make(chan []byte)
	go startMessageReader(conn, readChannel)
	for {
		message := decodeJsonToMessage(<-readChannel)
		if message.Content == "0" || message.Content == "1" {
			result = message
			break
		} else {
			fmt.Println("waiting...")
		}
	}
	conn.Close()
	return result
}

/**
 * openSignalChannel
 * conn: udp connection
 * oponentAddr: receiver address for udp
 * ment: good bye ment
 * When signal Ctrl+C inserted, print ment and exit process
 */
func openSignalChannel(conn net.PacketConn, oponentAddr net.Addr, ment string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		sig := <-sigs
		if sig.String() == "interrupt" {
			requestMessage := createP2PMessage(3, "")
			conn.WriteTo(requestMessage, oponentAddr)
			fmt.Println()
			if !gameOver {
				fmt.Println("You lose")
			}
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
 * content: response message content
 * myInfo: client information
 * return: json encoded byte array
 * create Message struct with option, content and encode it by json
 */
func createRequestMessage(content string, myInfo ClientInfo) []byte {
	msg := Message{Content: content, Info: myInfo}
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("json encode error: ", err)
	}
	return jsonMessage
}

/**
 * decodeJsonToMessage
 * buf: encoded json byte array
 * return: json decoded Message struct
 * get json byte array and decode it to Message struct
 */
func decodeJsonToMessage(buffer []byte) Message {
	var msg Message
	err := json.Unmarshal(buffer, &msg)
	if err != nil {
		log.Println("json err", err)
	}

	return msg
}

/**
 * decodeJsonToP2P
 * buf: encoded json byte array
 * return: json decoded P2PMessage struct
 * get json byte array and decode it to P2PMessage struct
 */
func decodeJsonToP2P(buffer []byte) P2PMessage {
	var msg P2PMessage
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
			conn.Close()
			return
		}
		readChannel <- buffer[:count]
	}
}

/**
* runTimer
* conn: udp conenction
* oponentAddr: receiver address for udp
* timeChannel: channel to receive request refresh time
* duration: duration for timer setting
* set timer for duration time. When new request come, refresh timer. When the timer is over, end omok
 */
func runTimer(conn net.PacketConn, oponentAddr net.Addr, timeChannel chan time.Time, duration time.Duration) {
	timer := time.NewTimer(time.Hour * 1000)
	for {
		select {
		case req_time := <-timeChannel:
			if gameOver || req_time.IsZero() {
				timer.Stop()
			} else {
				timer = time.NewTimer(duration)
			}
		case <-timer.C:
			if gameOver {
				return
			}
			gameOver = true
			conn.WriteTo(createP2PMessage(2, ""), oponentAddr)
			fmt.Println("You lose")
			return
		}
	}
}
