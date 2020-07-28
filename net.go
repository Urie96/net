package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	needHelp, isServer bool
	port               int
	ip                 string
)

func parse() {
	flag.BoolVar(&needHelp, "h", false, "this help")
	flag.BoolVar(&isServer, "s", false, "as a server")
	flag.IntVar(&port, "p", 80, "set port")
	flag.StringVar(&ip, "ip", "localhost", "set ip address")
	flag.Usage = func() {
		fmt.Println(`
Usage: tcp [OPTIONS] <port>

Options:`)
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()
	var err error
	port, err = strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Println("Error: invalid port")
		flag.Usage()
	}
}

func main() {
	parse()
	if needHelp {
		flag.Usage()
	}
	var network, addr string
	network = "tcp"
	addr = fmt.Sprintf("%s:%d", ip, port)
	if isServer {
		listen(network, addr)
	} else {
		dial(network, addr)
	}
}

func dial(network, addr string) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		log.Fatal("error net.Dial:", err)
	}
	log.Printf("dial %s succeed", addr)
	go loopRead(conn, nil)
	for {
		fmt.Println("what to send: (plz end with a new line fill with #)")
		contentToSend := getInput()
		go send(conn, contentToSend)
		time.Sleep(time.Second)
	}
}

func listen(network, addr string) {
	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Fatal("error net.Listen: ", err)
	}
	log.Printf("start listen %s at %s ...", listener.Addr().Network(), listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error net.Accept: ", err)
		}
		log.Printf("connection with %s is established", conn.RemoteAddr().String())
		go loopRead(conn, []byte("ok"))
	}
}

func loopRead(conn net.Conn, response []byte) {
	remoteAddr := conn.RemoteAddr().String()
	defer func() {
		log.Printf("connection with %s is closed", remoteAddr)
		conn.Close()
	}()
	var err error
	tips := fmt.Sprintf("data received from %s:\n", remoteAddr)
	for {
		var b []byte = make([]byte, 2*1024)
		_, err = conn.Read(b)
		if err != nil {
			log.Printf("error read data from %s(Error: %s)", remoteAddr, err.Error())
			break
		}
		log.Println(tips + string(b))
		if len(response) == 0 {
			continue
		}
		_, err := conn.Write(response)
		if err != nil {
			log.Printf("error write data to %s(Error: %s)", remoteAddr, err.Error())
		}
	}
}

func getInput() []byte {
	var str []byte
	isFirst := true
	appendStr := func(b []byte) {
		if isFirst {
			isFirst = false
		} else {
			str = append(str, "\n"...)
		}
		str = append(str, b...)
	}
	in := bufio.NewReader(os.Stdin)
	for {
		s, _, _ := in.ReadLine()
		if len(s) == 0 {
			appendStr(s)
			continue
		}
		isEnd := true
		for _, c := range s {
			if c != 35 {
				isEnd = false
				break
			}
		}
		if !isEnd {
			appendStr(s)
		} else {
			return str
		}
	}
}

func send(conn net.Conn, req []byte) {
	remoteAddr := conn.RemoteAddr().String()
	n, err := conn.Write(req)
	if err != nil {
		log.Printf("error write data to %s(Error: %s)", remoteAddr, err)
	}
	log.Printf("%d bytes sent succeed", n)
}
