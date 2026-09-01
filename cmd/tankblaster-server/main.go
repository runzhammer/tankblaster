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
	if value := os.Getenv("TANKBLASTER_SERVER_ADDRESS"); value != "" {
		cfg.Server.Address = value
	}
	if value := os.Getenv("TANKBLASTER_SERVER_PUBLIC_URL"); value != "" {
		cfg.Server.PublicURL = value
	}
	if value := os.Getenv("TANKBLASTER_SERVER_DB"); value != "" {
		cfg.Database.Path = value
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
