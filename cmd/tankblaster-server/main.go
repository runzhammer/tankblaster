package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/runzhammer/gamedemo/pkg/server"
)

func main() {
	configPath := flag.String("config", "config/server.example.yaml", "path to external YAML server config")
	flag.Parse()

	cfg, err := server.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	store, err := server.OpenStore(cfg)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(cfg, store).Run(ctx); err != nil {
		log.Fatal(err)
	}
}
