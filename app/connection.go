package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
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
	case '+': // simple string already complete
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
