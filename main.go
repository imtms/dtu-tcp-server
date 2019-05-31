package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	h         bool
	listenvar string
	redisvar  string
)

func usage() {
	log.Fatalf("Usage: dtuserver [-l listen_address] [-r redis_address] \n")
}

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&listenvar, "l", "0.0.0.0:8888", "listen_address")
	flag.StringVar(&redisvar, "r", "127.0.0.1:6379", "redis_address")
}

func main() {
	flag.Parse()

	if h {
		flag.Usage()
	}

	listener, err := net.Listen("tcp", listenvar)
	if err != nil {
		fmt.Printf("listen fail, err: %v\n", err)
		return
	}

	redisconn, err := redis.Dial("tcp", redisvar)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}
	defer redisconn.Close()

	fmt.Println("running on", listenvar)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept fail, err: %v\n", err)
			continue
		}
		fmt.Printf("new Connection： con=%v ip=%v\n", conn, conn.RemoteAddr().String())
		go process(conn, redisconn)
	}
}

func process(conn net.Conn, redisconn redis.Conn) {
	defer conn.Close()
	for {
		var buf [128]byte
		_, err := conn.Read(buf[:])
		if err != nil {
			fmt.Printf("read from connect failed, err: %v\n", err)
			break
		}
		if string(buf[0:2]) != "86" {
			fmt.Println("error header")
			break
		}
		imei := string(buf[0:15])
		encodedStr := hex.EncodeToString(buf[15:26])
		_, err = redisconn.Do("LPUSH", imei, encodedStr+strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			fmt.Println("redis set failed:", err)
			break
		}
	}
}
