package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
func readArray(n int, r *bufio.Reader) ([]string, error) {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		prefix, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		switch prefix {
		case '$': // bulk string
			lenLine, _ := r.ReadString('\n')
			size, err := strconv.Atoi(strings.TrimSuffix(lenLine, "\r\n"))
			if err != nil {
				return nil, fmt.Errorf("bulk len: %v", err)
			}
			data := make([]byte, size+2) // +2 for CRLF
			if _, err := io.ReadFull(r, data); err != nil {
				return nil, err
			}
			out = append(out, string(data[:size]))
		case ':': // integer
			intLine, _ := r.ReadString('\n')
			out = append(out, strings.TrimSuffix(intLine, "\r\n"))
		default:
			return nil, fmt.Errorf("bad prefix %q", prefix)
		}
	}
	return out, nil
}

func respArrays(arr []string) string {
	var resp string = fmt.Sprintf("*%d\r\n", len(arr))

	for _, str := range arr {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
	}

	return resp
}
