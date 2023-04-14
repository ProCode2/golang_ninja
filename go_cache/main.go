package main

import (
	"flag"
	"golang_ninja/go_cache/cache"
	"log"
	"net"
	"time"
)

func main() {
	listenAddress := flag.String("listenaddr", ":3000", "listen address of the server")
	leaderAddress := flag.String("leaderaddr", "", "leader address of the server")
	flag.Parse()

	opts := ServerOpts{
		ListenAddress: *listenAddress,
		LeaderAddress: *leaderAddress,
		IsLeader:      len(*leaderAddress) == 0,
	}

	go func() {
		time.Sleep(time.Second * 2)
		conn, err := net.Dial("tcp", opts.ListenAddress)
		if err != nil {
			log.Printf("dial error: %v", err)
			return
		}

		conn.Write([]byte("SET FOO BAR 250000000"))

		time.Sleep(time.Second * 2)

		conn.Write([]byte("GET FOO"))
		buf := make([]byte, 1000)
		n, _ := conn.Read(buf)

		log.Println(string(buf[:n]))
	}()

	server := NewServer(opts, cache.NewCache())
	server.Start()
}
