package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func handleCommands(firstLine string, r *bufio.Reader) (string, error) {
	if len(firstLine) == 0 {
		return "", fmt.Errorf("empty request")
	}

	switch firstLine[0] {
	case '*': // array
		num, err := strconv.Atoi(firstLine[1:])
		if err != nil {
			return "", fmt.Errorf("bad array length: %v", err)
		}
		arr, err := readArray(num, r)
		if err != nil {
			return "", err
		}
		return MatchCommand(arr)
	case '+':
		return firstLine + "\r\n", nil
	default:
		return "", fmt.Errorf("unexpected type byte %q", firstLine[0])
	}
}

func EventLoop(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("read:", err)
			}
			return
		}
		resp, err := handleCommands(strings.TrimSuffix(string(line), "\r\n"), r)
		if err != nil {
			conn.Write([]byte("-ERR " + err.Error() + "\r\n"))
			return
		}
		conn.Write([]byte(resp))
	}
}

func handleReplicaOf(conn net.Conn, port string) {
	_, err := conn.Write([]byte(respArrays([]string{"PING"})))
	if err != nil {
		fmt.Println("Failed to write ping to master.\nClosing connection...")
		conn.Close()
		os.Exit(1)
	}
	r := bufio.NewReader(conn)
	line, err := r.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("read:", err)
		}
		return
	}
	if strings.Compare(string(line), "+PONG\r\n") != 0 {
		fmt.Println("Master didn't responsed with PONG")
		conn.Close()
		os.Exit(1)
	}
	_, err = conn.Write([]byte(respArrays([]string{"REPLCONF", "listening-port", port})))
	if err != nil {
		fmt.Println("Failed to write REPLCONF -port to master.\nClosing connection...")
		conn.Close()
		os.Exit(1)
	}
	line, err = r.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("read:", err)
		}
		return
	}
	if strings.Compare(string(line), "+OK\r\n") != 0 {
		fmt.Println("Master didn't responsed with +OK")
		conn.Close()
		os.Exit(1)
	}
	_, err = conn.Write([]byte(respArrays([]string{"REPLCONF", "capa", "psync2"})))
	if err != nil {
		fmt.Println("Failed to write REPLCONF -capa psync2 to master.\nClosing connection...")
		conn.Close()
		os.Exit(1)
	}
	line, err = r.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("read:", err)
		}
		return
	}
	if strings.Compare(string(line), "+OK\r\n") != 0 {
		fmt.Println("Master didn't responsed with +OK")
		conn.Close()
		os.Exit(1)
	}
	_, err = conn.Write([]byte(respArrays([]string{"PSYNC", "?", "-1"})))
	if err != nil {
		fmt.Println("Failed to write REPLCONF -psync to master.\nClosing connection...")
		conn.Close()
		os.Exit(1)
	}
	line, err = r.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("read:", err)
		}
		return
	}
	EventLoop(conn)
}
