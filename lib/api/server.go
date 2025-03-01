package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
)

type Server struct {
	server *http.Server
}

func NewServer(ctx context.Context) (*Server, error) {
	svr := &http.Server{}
	if err := http2.ConfigureServer(svr, nil); err != nil {
		return nil, err
	}

	context.AfterFunc(ctx, func() {
		if err := svr.Shutdown(context.TODO()); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown server")
		} else {
			log.Info().Msg("Server shutdown")
		}
	})

	return &Server{
		server: svr,
	}, nil
}

func (s *Server) RegisterHandler(path string, handler http.Handler) {
	http.Handle(path, handler)
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	list, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	context.AfterFunc(ctx, func() {
		list.Close()
	})

	proxyListener := &proxyproto.Listener{Listener: list}

	if err := s.server.Serve(proxyListener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
