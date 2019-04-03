package server

import (
	"context"
	"net"

	"net/http"

	"sync"

	"time"

	"errors"

	"github.com/molon/pkg/server/gateway"
	"github.com/molon/pkg/server/health"
	"google.golang.org/grpc"
)

var (
	// ErrInitializeTimeout is returned when an InitializerFunc takes too long to finish during Server.Serve
	ErrInitializeTimeout = errors.New("initialization timed out")
	// DefaultInitializerTimeout is the reasonable default amount of time one would expect initialization to take in the
	// worst case
	DefaultInitializerTimeout = time.Minute
)

// Server is a wrapper struct that will allow you to stand up your GPRC server, HTTP server and health checks within
// the same struct. The recommended way to initialize this is with the NewServer function.
type Server struct {
	mu      sync.Mutex
	stopped bool
	serving bool

	initializers      []InitializerFunc
	initializeTimeout time.Duration
	registrars        []func(mux *http.ServeMux) error

	// GRPCServer will be started whenever this is served
	GRPCServer *grpc.Server

	// HTTPServer will be started whenever this is served
	HTTPServer *http.Server
}

// Option is a functional option for creating a Server
type Option func(*Server) error

// InitializerFunc is a handler that can be passed into WithInitializer to be executed prior to serving
type InitializerFunc func(context.Context) error

// NewServer creates a Server from the given options. All options are processed in the order they are declared.
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		initializeTimeout: DefaultInitializerTimeout,
		HTTPServer:        &http.Server{},
		registrars:        []func(mux *http.ServeMux) error{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	for _, register := range s.registrars {
		if err := register(mux); err != nil {
			return nil, err
		}
	}
	s.HTTPServer.Handler = mux

	return s, nil
}

// WithInitializerTimeout set the duration initialization will wait before halting and returning an error
func WithInitializerTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.initializeTimeout = timeout
		return nil
	}
}

// WithInitializer adds an initialization function that will get called prior to serving.
func WithInitializer(initializerFunc InitializerFunc) Option {
	return func(s *Server) error {
		s.initializers = append(s.initializers, initializerFunc)
		return nil
	}
}

// WithGRPCServer adds the given GRPC server to this server. There can only be one GRPC server within a given instance,
// so multiple calls with this option will overwrite the previous ones.
func WithGRPCServer(grpcServer *grpc.Server) Option {
	return func(s *Server) error {
		s.GRPCServer = grpcServer
		return nil
	}
}

// WithHTTPHandler registers the given http handler to this server by registering the pattern at the root of the http server
func WithHTTPHandler(pattern string, handler http.Handler) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			mux.Handle(pattern, handler)
			return nil
		})
		return nil
	}
}

// WithHealthChecker registers the given health checker with this server by registering its endpoints at the root of the
// http server.
func WithHealthChecker(checker health.Checker) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			checker.RegisterHandler(mux)
			return nil
		})
		return nil
	}
}

// WithGateway registers the given gateway options with this server
func WithGateway(options ...gateway.Option) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			_, err := gateway.NewGateway(append(options, gateway.WithMux(mux))...)
			return err
		})
		return nil
	}
}

// Serve invokes all initializers then serves on the given listeners.
//
// If a listener is left blank, then that particular part will not be served.
//
// If a listener is specified for a part that doesn't have a corresponding server, then an error will be returned. This
// can happen, for instance, whenever a gRPC listener is provided but no gRPC server was set or no option was passed
// into NewServer.
//
// If both listeners are nil, then an error is returned
func (s *Server) Serve(grpcL, httpL net.Listener) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return errors.New("server is already stopped")
	}
	if s.serving {
		s.mu.Unlock()
		return errors.New("server is already in service")
	}
	s.serving = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.serving = false
		s.mu.Unlock()
	}()

	if grpcL == nil && httpL == nil {
		return errors.New("both grpcL and httpL are nil")
	}

	if err := s.initialize(); err != nil {
		return err
	}
	errC := make(chan error, 1)

	if httpL != nil {
		if s.HTTPServer == nil {
			return errors.New("httpL is specified, but no HTTPServer is provided")
		}
		go func() { errC <- s.HTTPServer.Serve(httpL) }()
	} else {
		s.HTTPServer = nil
	}

	if grpcL != nil {
		if s.GRPCServer == nil {
			return errors.New("grpcL is specified, but no GRPCServer is provided")
		}
		go func() { errC <- s.GRPCServer.Serve(grpcL) }()
	} else {
		s.GRPCServer = nil
	}
	defer s.Stop()
	return <-errC
}

// Stop immediately terminates the grpc and http servers
func (s *Server) Stop() error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	s.mu.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(2)
	errC := make(chan error, 1)
	go func() {
		defer wg.Done()

		if s.GRPCServer != nil {
			stopped := make(chan struct{})
			go func() {
				s.GRPCServer.GracefulStop()
				close(stopped)
			}()
			// 先尝试GracefulStop，10秒还不成，再强制Stop
			t := time.NewTimer(10 * time.Second)
			select {
			case <-t.C:
				s.GRPCServer.Stop()
			case <-stopped:
				t.Stop()
			}
		}
	}()
	go func() {
		defer wg.Done()
		if s.HTTPServer != nil {
			if err := s.HTTPServer.Close(); err != nil {
				errC <- err
			}
		}
	}()

	go func() {
		wg.Wait()
		close(errC)
	}()

	return <-errC
}

func (s *Server) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.initializeTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(len(s.initializers))
	errC := make(chan error, len(s.initializers))
	for _, initFunc := range s.initializers {
		go func(init InitializerFunc) {
			defer wg.Done()
			if err := init(ctx); err != nil {
				errC <- err
			}
		}(initFunc)
	}

	go func() {
		wg.Wait()
		close(errC)
	}()

	t := time.NewTimer(s.initializeTimeout)
	select {
	case err := <-errC:
		t.Stop() // for gc
		return err
	case <-t.C:
		return ErrInitializeTimeout
	}
}

func GracefulStop(svcs ...*Server) {
	wg := sync.WaitGroup{}
	wg.Add(len(svcs))
	for _, svc := range svcs {
		go func(svc *Server) {
			defer wg.Done()
			svc.Stop()
		}(svc)
	}
	wg.Wait()
}
