package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Server struct {
	context  context.Context
	server   http.Server
	listener *net.TCPListener
}

func (s *Server) Listen(ctx context.Context, listenAddress string) (*net.TCPAddr, error) {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", listenAddress)
	if err != nil {
		return nil, err
	}
	s.context = ctx
	s.listener = listener.(*net.TCPListener)
	return s.Address(), nil
}

func (s *Server) Serve(handler http.Handler) error {
	if s.listener == nil {
		return errors.New("server must be listening first")
	}
	s.server = http.Server{
		Addr:        s.listener.Addr().String(),
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return s.context },
	}
	return s.server.Serve(s.listener)
}

func IsClosedError(err error) bool {
	return errors.Is(err, http.ErrServerClosed) || errors.Is(err, context.Canceled)
}

func (s *Server) Context() context.Context {
	return s.context
}

func (s *Server) Address() *net.TCPAddr {
	addr := s.listener.Addr().(*net.TCPAddr)
	addrCopy := *addr
	return &addrCopy
}

func (s *Server) AddressString() string {
	addr := s.Address()
	if addr == nil {
		return ""
	}
	return addr.AddrPort().String()
}

func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
