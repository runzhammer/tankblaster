package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/runzhammer/tankblaster/pkg/buildinfo"
	"github.com/runzhammer/tankblaster/pkg/server"
)

func main() {
	configPath := flag.String("config", "config/server.example.yaml", "path to external YAML server config")
	showVersion := flag.Bool("version", false, "print version information")
	flag.Parse()
	if *showVersion {
		fmt.Println(buildinfo.String())
		return
	}

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
	if value := os.Getenv("TANKBLASTER_SERVER_ALLOWED_ORIGINS"); value != "" {
		cfg.Server.AllowedOrigins = splitCSV(value)
	}
	if value := os.Getenv("TANKBLASTER_SERVER_MAX_CONNECTIONS"); value != "" {
		cfg.Server.MaxConnections = parsePositiveInt("TANKBLASTER_SERVER_MAX_CONNECTIONS", value)
	}
	if value := os.Getenv("TANKBLASTER_SERVER_MAX_SESSIONS"); value != "" {
		cfg.Server.MaxSessions = parsePositiveInt("TANKBLASTER_SERVER_MAX_SESSIONS", value)
	}
	if value := os.Getenv("TANKBLASTER_SERVER_MAX_QUEUE_LENGTH"); value != "" {
		cfg.Server.MaxQueueLength = parsePositiveInt("TANKBLASTER_SERVER_MAX_QUEUE_LENGTH", value)
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

func splitCSV(value string) []string {
	fields := strings.Split(value, ",")
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			out = append(out, field)
		}
	}
	return out
}

func parsePositiveInt(name, value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n <= 0 {
		log.Fatalf("%s must be a positive integer", name)
	}
	return n
}
