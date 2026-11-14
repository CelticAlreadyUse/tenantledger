// Package app composes modules and shared infrastructure into one application.
// Cross-module dependency wiring belongs here, outside module business logic.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/samber/do/v2"
	"github.com/samber/oops"

	"github.com/CelticAlreadyUse/tenantledger/internal/platform/httpserver"
)

// Application owns the dependency container and the process-level services.
// Modules receive explicit constructor dependencies rather than the injector.
type Application struct {
	injector *do.RootScope
	server   *http.Server
}

// New registers and resolves the application dependencies before serving traffic.
func New(cfg Config) (*Application, error) {
	injector := do.New()
	do.ProvideValue(injector, cfg)
	do.Provide(injector, func(do.Injector) (http.Handler, error) {
		return NewHandler(), nil
	})
	do.Provide(injector, provideHTTPServer)

	server, err := do.Invoke[*http.Server](injector)
	if err != nil {
		injector.Shutdown()
		return nil, oops.In("app").Code("dependency_init_failed").Wrap(err)
	}
	return &Application{injector: injector, server: server}, nil
}

func provideHTTPServer(i do.Injector) (*http.Server, error) {
	cfg, err := do.Invoke[Config](i)
	if err != nil {
		return nil, err
	}
	handler, err := do.Invoke[http.Handler](i)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr: cfg.HTTPAddr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}, nil
}

// Run serves HTTP until cancellation or a listener failure. An Application is
// single-use: its instantiated dependencies are shut down before Run returns.
func (a *Application) Run(ctx context.Context) (runErr error) {
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// http.Server implements do's Shutdown(context.Context) error contract.
		report := a.injector.ShutdownWithContext(shutdownCtx)
		if !report.Succeed {
			// Close remaining connections if graceful shutdown exceeded its deadline.
			_ = a.server.Close()
			runErr = errors.Join(runErr, oops.In("app").Code("shutdown_failed").Wrap(report))
		}
	}()

	listener, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return oops.In("httpserver").Code("listen_failed").
			With("address", a.server.Addr).Wrap(err)
	}
	slog.Info("API listening", "addr", listener.Addr().String())
	serveErr := make(chan error, 1)
	go func() { serveErr <- a.server.Serve(listener) }()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return oops.In("httpserver").Code("serve_failed").Wrap(err)
	}
}

// NewHandler builds the HTTP application. Module routes will be registered here
// through each module's exported facade as their features are implemented.
func NewHandler() http.Handler {
	return httpserver.NewRouter()
}
