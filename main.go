package main

import (
	"bufio"
	"github.com/cdgn-coding/redis-compatible-challenge/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/resp"
	"log"
	"net"
	"os"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

var parser = resp.RespParser{}

var serializer = resp.RespSerializer{}

var eng = engine.NewEngine()

func handleClient(conn net.Conn) {
	var serialized []byte
	for {
		scanner := bufio.NewScanner(conn)
		payload, err := parser.ParseWithScanner(scanner)
		if err != nil {
			logger.Println(err)
			return
		}

		// Process payload
		res, err := eng.Process(payload)

		// Report engine errors
		if err != nil {
			logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized)
			return
		}

		// Serialize response
		serialized, err = serializer.Serialize(res)

		// Report serialization errors
		if err != nil {
			logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized)
			return
		}

		// Write response
		_, err = conn.Write(serialized)
	}

	defer conn.Close()
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		logger.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Println(err)
			continue
		}

		logger.Printf("accepted connection from %s", conn.RemoteAddr())
		go handleClient(conn)
	}

	defer listener.Close()
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
