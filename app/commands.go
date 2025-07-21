package main

import (
	"fmt"
	"strconv"
	"strings"
)

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
			mp.set(arr[1], arr[2], -1)
		} else if len(arr) == 5 && strings.ToLower(arr[3]) == "px" {
			n, err := strconv.Atoi(arr[4])
			if err != nil {
				return "-Err px not a valid number\r\n", err
			}
			mp.set(arr[1], arr[2], n)
		}
		return "+OK\r\n", nil
	case "get":
		val, ok := mp.get(arr[1])
		if !ok {
			return "$-1\r\n", nil
		}
		return respBulkString([]string{val}), nil
	case "config":
		if len(arr) >= 3 && strings.ToLower(arr[1]) == "get" {
			switch strings.ToLower(arr[2]) {
			case "dir":
				return respArrays([]string{"dir", config.Directory}), nil
			case "dbfilename":
				return respArrays([]string{"dbfilename", config.dbFileName}), nil
			default:
				return "-Arguments too short\r\n", nil
			}
		}
	case "keys":
		if len(arr) == 2 {
			pattern := arr[1]
			Keys := mp.keys(pattern)
			return respArrays(Keys), nil
		}
		return "-Arguments too short\r\n", nil
	case "info":
		res := ""
		switch repli.role {
		case 0:
			res = "master"
		case 1:
			res = "slave"
		default:
			res = ""
		}
		return respBulkString([]string{fmt.Sprintf("role:%s", res), "connected_slaves:0",
			"master_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
			"master_repl_offset:0"}), nil
	case "replconf":
		return "+OK\r\n", nil
	case "psync":
		return fmt.Sprintf("+FULLRESYNC %s 0\r\n", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"), nil
	default:
		return "-ERR unknown command\r\n", nil
	}
	return "-ERR unknown command\r\n", nil
}

func respBulkString(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	payload := b.String()
	return "$" + strconv.Itoa(len(payload)) + "\r\n" + payload + "\r\n"
}
