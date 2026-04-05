package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/swilcox/led-kurokku-go/internal/admin"
	"github.com/swilcox/led-kurokku-go/version"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	storePath := flag.String("store", defaultStorePath(), "path to instance store JSON file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("kurokku-admin", version.String())
		return
	}

	store := admin.NewStore(*storePath)
	if err := store.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load store: %v\n", err)
		os.Exit(1)
	}

	srv := admin.NewServer(store)

	slog.Info("kurokku-admin starting", "version", version.String(), "addr", *addr, "store", *storePath)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func defaultStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".kurokku-admin.json"
	}
	return filepath.Join(home, ".kurokku-admin.json")
}
