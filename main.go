package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	tcp "github.com/jeffque/teecp/teecp"
)

func main() {
	port := ":6667"
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		serverTeecp(port)
	} else {
		listenerTeecp(port)
	}
}

func listenerTeecp(port string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost%s", port))
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Could not open socket to port %s\n", port))
		os.Exit(1)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		txt, err := reader.ReadString('\n')
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Some error while reading [%s], closing stream\n", err.Error()))
			os.Exit(1)
		}

		fmt.Println(txt[:len(txt)-1])
	}
}

func serverTeecp(port string) {
	var teecp tcp.TeeCPList = tcp.TeeCPList{}

	tcp.Attach(&teecp, func(msg string) bool {
		fmt.Println(msg)
		return true
	})

	ln, err := net.Listen("tcp", port)

	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Could not open socket to port %s\n", port))
		os.Exit(1)
	}
	defer ln.Close()
	go acceptNewConns(ln, &teecp)

	// tcp.Attach(&teecp)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		txt := scanner.Text()
		teecp.Broadcast(txt)
	}
}

func acceptNewConns(ln net.Listener, teecp *tcp.TeeCPList) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("tried to connect but failed %s\n", err.Error()))
			continue
		}
		go handleConnection(conn, teecp)
	}
}

func handleConnection(conn net.Conn, teecp *tcp.TeeCPList) {
	tcp.Attach(teecp, func(msg string) bool {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			conn.Close()
			return false
		}
		_, err2 := conn.Write([]byte("\n"))
		if err2 != nil {
			conn.Close()
			return false
		}
		return true
	})
}
