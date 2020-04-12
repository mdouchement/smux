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
		defer l.Close()

		go conn.Listen()

		go func() {
			for {
				stream, err := conn.Accept()
				if err != nil {
					break
				}

				go func() {
					go stream.Poll()

					var buf bytes.Buffer
					out := make([]byte, 512)
					for {
						n, err := stream.Read(out)
						if err == io.EOF {
							break
						}
						buf.Write(out[:n])
					}

					var b bytes.Buffer
					w := bufio.NewWriter(&b)
					s.Handler.Serve(w, bytes.NewReader(buf.Bytes()))
					w.Flush()
					stream.WriteOnce(b.Bytes())
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
