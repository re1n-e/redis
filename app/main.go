package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

type Config struct {
	Directory  string
	dbFileName string
}

var config Config
var rdb = &RDB{}
var mp = &MP{}

func main() {
	fmt.Println("Logs from your program will appear here!")

	dir := flag.String("dir", "", "Directory to store the database")
	dbfilename := flag.String("dbfilename", "", "Database file name")
	flag.Parse()

	config = Config{
		Directory:  *dir,
		dbFileName: *dbfilename,
	}

	path := fmt.Sprintf("%s/%s", config.Directory, config.dbFileName)
	if config.Directory != "" && config.dbFileName != "" {
		if _, err := os.Stat(path); err == nil {
			buffer, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading RDB file:", err)
				os.Exit(1)
			}

			loadedMP, err := LoadRDBToMP(buffer)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to load RDB into memory:", err)
				os.Exit(1)
			}

			mp = loadedMP
		} else if os.IsNotExist(err) {
			fmt.Println("RDB file does not exist yet, starting with empty memory")
			// starting with an empty map
		} else {
			fmt.Fprintln(os.Stderr, "Error checking RDB file:", err)
			os.Exit(1)
		}
	}
	// Bind to TCP port 6379
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to bind to port 6379:", err)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error accepting connection:", err)
			os.Exit(1)
		}
		go EventLoop(conn)
	}
}
