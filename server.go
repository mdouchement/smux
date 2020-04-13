package smux

import (
	"bufio"
	"bytes"
	"io"
)

type Server struct {
	Network string
	Address string
	Handler Handler

	listener io.Closer
}

func (s *Server) ListenAndServe() error {
	l, err := Listen(s.Network, s.Address)
	if err != nil {
		return err
	}
	s.listener = l

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		go conn.Listen()

		go func() {
			for {
				stream, err := conn.Accept()
				if err != nil {
					break
				}

				go func() {
					go stream.Poll()

					var w bytes.Buffer
					s.Handler.Serve(&w, bufio.NewReader(stream))
					stream.WriteOnce(w.Bytes())
				}()
			}
		}()
	}
}

// Shutdown stops the server.
func (s *Server) Shutdown() {
	s.listener.Close()
}

type Handler interface {
	Serve(io.Writer, io.Reader)
}

type HandlerFunc func(io.Writer, io.Reader)

func (f HandlerFunc) Serve(w io.Writer, r io.Reader) {
	f(w, r)
}
