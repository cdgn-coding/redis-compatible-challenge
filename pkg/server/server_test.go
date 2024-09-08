package server

import (
	"bufio"
	"context"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/resp"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestServer_SET_GET(t *testing.T) {
	s := &Server{
		eng:    engine.NewEngine(),
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan struct{})
	go s.StartServer(ctx, "3000", ready)
	<-ready

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	conn.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))
	<-time.After(1 * time.Second)

	parser := resp.RespParser{}

	res, err := parser.ParseWithScanner(scanner)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(res).Kind() != reflect.String || res.(string) != "OK" {
		t.Fatal("Expected SET to return string OK")
	}

	conn.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"))
	<-time.After(1 * time.Second)

	res, err = parser.ParseWithScanner(scanner)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(res).Kind() != reflect.String || res.(string) != "value" {
		t.Fatal("Expected GET key to return string value")
	}

	cancel()
}
