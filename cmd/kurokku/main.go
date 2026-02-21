package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/engine"
	"github.com/swilcox/led-kurokku-go/redis"
)

func main() {
	displayOverride := flag.String("display", "", "display type override (terminal, max7219, tm1637, ht16k33, terminal_seg7, terminal_seg14)")
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	// Initialize Redis (optional).
	rds, err := redis.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "redis config error: %v\n", err)
		os.Exit(1)
	}
	if rds != nil {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := rds.Ping(pingCtx); err != nil {
			log.Printf("redis ping failed, running without redis: %v", err)
			rds = nil
		} else {
			log.Println("redis connected")
			defer rds.Close()
		}
		pingCancel()
	}

	// Load config: try Redis first, fall back to file.
	var cfg *config.Config
	if rds != nil {
		loadCtx, loadCancel := context.WithTimeout(context.Background(), 5*time.Second)
		redisCfg, found, fetchErr := rds.FetchConfig(loadCtx)
		loadCancel()
		if fetchErr != nil {
			log.Printf("redis config fetch error, falling back to file: %v", fetchErr)
		} else if found {
			cfg = redisCfg
			log.Println("config loaded from Redis")
		}
	}
	if cfg == nil {
		cfg, err = config.Load(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config error: %v\n", err)
			os.Exit(1)
		}
	}

	// CLI -display flag overrides config display.type
	if *displayOverride != "" {
		cfg.Display.Type = config.DisplayType(*displayOverride)
	}
	// Default to terminal if not set
	if cfg.Display.Type == "" {
		cfg.Display.Type = config.DisplayTerminal
	}

	disp, err := createDisplay(cfg.Display)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := disp.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "display init failed: %v\n", err)
		os.Exit(1)
	}
	defer disp.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Subscribe to config key changes for hot-reload (nil when Redis absent).
	var configCh <-chan struct{}
	if rds != nil {
		if ch, err := rds.SubscribeConfig(ctx); err != nil {
			log.Printf("config subscribe failed: %v", err)
		} else {
			configCh = ch
		}
	}

	for {
		engCtx, engCancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() { done <- engine.New(disp, cfg, rds).Run(engCtx) }()

		select {
		case err := <-done:
			engCancel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "engine error: %v\n", err)
				os.Exit(1)
			}
			disp.Clear()
			return

		case <-configCh:
			engCancel()
			<-done
			reloadCtx, reloadCancel := context.WithTimeout(context.Background(), 5*time.Second)
			newCfg, found, fetchErr := rds.FetchConfig(reloadCtx)
			reloadCancel()
			if fetchErr != nil {
				log.Printf("config reload error: %v", fetchErr)
			} else if found {
				cfg = newCfg
				log.Println("config reloaded from Redis")
			}
			// loop → restart engine with (possibly updated) cfg

		case <-ctx.Done():
			engCancel()
			<-done
			disp.Clear()
			return
		}
	}
}

func createDisplay(dc config.DisplayConfig) (display.Display, error) {
	switch dc.Type {
	case config.DisplayTerminal:
		return display.NewTerminal(os.Stdout), nil
	case config.DisplayMAX7219:
		return display.NewMAX7219(""), nil
	case config.DisplayTerminalSeg7:
		return display.NewTerminalSegment(os.Stdout, display.Segment7), nil
	case config.DisplayTerminalSeg14:
		return display.NewTerminalSegment(os.Stdout, display.Segment14), nil
	case config.DisplayTM1637:
		return display.NewTM1637(dc.ClkPin, dc.DioPin), nil
	case config.DisplayHT16K33:
		addr := dc.I2CAddr
		if addr == 0 {
			addr = 0x70
		}
		return display.NewHT16K33(dc.I2CBus, addr, dc.Layout), nil
	default:
		return nil, fmt.Errorf("unknown display type: %s", dc.Type)
	}
}
