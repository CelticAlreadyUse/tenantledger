package app

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/samber/oops"
)

func TestRunPreservesListenerError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	application, err := New(Config{HTTPAddr: listener.Addr().String()})
	if err != nil {
		t.Fatal(err)
	}
	shutdown := make(chan struct{})
	application.server.RegisterOnShutdown(func() { close(shutdown) })
	err = application.Run(context.Background())
	var networkErr *net.OpError
	if !errors.As(err, &networkErr) {
		t.Fatalf("expected original network error, got %v", err)
	}
	structured, ok := oops.AsOops(err)
	if !ok || structured.Code() != "listen_failed" {
		t.Fatalf("expected structured listen_failed error, got %v", err)
	}
	if structured.Context()["address"] != listener.Addr().String() {
		t.Fatal("listener address metadata missing")
	}
	if len(structured.StackFrames()) == 0 {
		t.Fatal("error stack trace missing")
	}
	select {
	case <-shutdown:
	case <-time.After(time.Second):
		t.Fatal("container did not shut down HTTP server after listener failure")
	}
}

func TestRunShutsDownOnCancellation(t *testing.T) {
	application, err := New(Config{HTTPAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	shutdown := make(chan struct{})
	application.server.RegisterOnShutdown(func() { close(shutdown) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := application.Run(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-shutdown:
	case <-time.After(time.Second):
		t.Fatal("container did not shut down HTTP server on cancellation")
	}
}
