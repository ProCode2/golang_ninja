package main

import (
	"context"
	"fmt"
	"golang_ninja/go_cache/cache"
	"log"
	"net"
)

type ServerOpts struct {
	ListenAddress string
	LeaderAddress string
	IsLeader      bool
}

type Server struct {
	ServerOpts
	followers map[net.Conn]struct{}
	cache     cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
		// only allocate this when we are the leader
		followers: make(map[net.Conn]struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen error: %s", err)
	}

	log.Printf("Server starting on port %s.\n", s.ListenAddress)

	if !s.IsLeader {
		go func() {
			conn, err := net.Dial("tcp", s.LeaderAddress)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Connected with leader")
			s.handleConn(conn)
		}()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept Error: %s", err)
			continue
		}

		go s.handleConn(conn)

	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	buf := make([]byte, 2048)

	if s.IsLeader {
		s.followers[conn] = struct{}{}
	}

	log.Println("connection made: ", conn.RemoteAddr())
	for {
		n, err := conn.Read(buf)

		if err != nil {
			log.Printf("Conn Read Error: %s", err)
			break
		}

		s.handleCommmand(conn, buf[:n])
	}

}

func (s *Server) handleCommmand(conn net.Conn, rawCmd []byte) {
	msg, err := parseMessage(rawCmd)

	if err != nil {
		log.Println("failed to parse command", err)
		return
	}

	log.Println("Received Command: ", msg.Cmd)

	switch msg.Cmd {
	case CMDSet:
		err = s.handleSetCmd(conn, msg)
	case CMDGet:
		err = s.handleGetCmd(conn, msg)
	}

	if err != nil {
		log.Println("failed to handle command: ", err)
		return
	}
}

func (s *Server) handleSetCmd(conn net.Conn, msg *Message) error {
	if err := s.cache.Set(msg.Key, msg.Value, msg.TTL); err != nil {
		return err
	}

	go s.sendToFollowers(context.TODO(), msg)
	return nil
}

func (s *Server) handleGetCmd(conn net.Conn, msg *Message) error {
	value, err := s.cache.Get(msg.Key)
	if err != nil {
		return err
	}
	_, err = conn.Write(value)

	return err
}

func (s *Server) sendToFollowers(ctx context.Context, msg *Message) error {
	for conn := range s.followers {
		log.Println("forwarding key to followers")
		_, err := conn.Write(msg.toBytes())
		if err != nil {
			log.Println("write to followers error: ", err)
			continue
		}
	}
	return nil
}
