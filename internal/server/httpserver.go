package server

import (
	"errors"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/Ciobi0212/httpfromtcp/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	Router   *response.Router
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := ":" + strconv.Itoa(port)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	server := Server{
		Port:     port,
		Listener: listener,
		isClosed: atomic.Bool{},
		Router:   response.NewRouter(),
	}

	go func() {
		server.listen()
	}()

	return &server, nil
}

func (s *Server) Close() error {
	s.isClosed.Swap(true)
	return s.Listener.Close()
}

func (s *Server) listen() {
	for {
		log.Println("Listening for connection")

		conn, err := s.Listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Listener closed, stopping accept loop.")
				return
			}

			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Println("Connection received")

		responseWriter := response.ResponseWriter{
			Conn: conn,
		}

		go s.Router.Handle(responseWriter)
	}
}
