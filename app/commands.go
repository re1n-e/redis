package main

import (
	"fmt"
	"strconv"
	"strings"
)

var db = &DB{}

func MatchCommand(arr []string) (string, error) {
	if len(arr) == 0 {
		return "", fmt.Errorf("empty array")
	}
	switch strings.ToLower(arr[0]) {
	case "ping":
		return "+PONG\r\n", nil
	case "echo":
		return respBulkString(arr[1:]), nil
	case "set":
		if len(arr) == 3 {
			db.set(arr[1], arr[2], -1)
		} else if len(arr) == 5 && strings.ToLower(arr[3]) == "px" {
			n, err := strconv.Atoi(arr[4])
			if err != nil {
				return "-Err px not a valid number\r\n", err
			}
			db.set(arr[1], arr[2], n)
		}
		return "+OK\r\n", nil
	case "get":
		val, ok := db.get(arr[1])
		if !ok {
			return "$-1\r\n", nil
		}
		return respBulkString([]string{val}), nil
	default:
		return "-ERR unknown command\r\n", nil
	}
}

func respBulkString(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	payload := b.String()
	return "$" + strconv.Itoa(len(payload)) + "\r\n" + payload + "\r\n"
}
