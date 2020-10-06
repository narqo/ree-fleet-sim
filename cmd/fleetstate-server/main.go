package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/fleetstate"
	"github.com/narqo/ree-fleet-sim/internal/middleware"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("", flag.ExitOnError)

	var (
		httpAddr        string
		shutdownTimeout time.Duration
	)
	flags.StringVar(&httpAddr, "http-addr", "127.0.0.1:10080", "address to listen on")
	flags.DurationVar(&shutdownTimeout, "http-shutdown-timeout", 5*time.Second, "server shutdown timeout")

	if err := flags.Parse(args); err != nil {
		return err
	}

	store := fleetstate.NewMemStore()
	vh := fleetstate.NewVehicleHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/vehicle/", http.StripPrefix("/vehicle", vh.Handler()))

	handler := middleware.LoggingHandler(os.Stdout, mux)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	errs := make(chan error, 1)
	go func() {
		log.Printf("listening: addr %s", server.Addr)
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("exiting...")
	case err := <-errs:
		return err
	}

	// create new context because top-most one is already canceled
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return server.Shutdown(ctx)
}
