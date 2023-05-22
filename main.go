package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
)

func sygnals(s shutdowner, cfg *Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		sig := <-c
		switch sig {
		case os.Interrupt:
			//handle SIGINT
			logrus.Printf("shutdown")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()

			err := s.Shutdown(ctx)
			if err != nil {
				logrus.Printf("shutdown error: %s", err)
				os.Exit(1)
			}

			cfg.wg.Wait()
			os.Exit(0)
		}
	}()
}

func main() {
	cfgFile := flag.String("c", "config.yml", "config file")
	flag.Parse()

	cfg, err := LoadCfg(*cfgFile)
	if err != nil {
		logrus.Fatalf("cfg read error: %s", err)
	}

	logrus.SetLevel(cfg.LogLevel)

	s := newServer(cfg)
	sygnals(s, cfg)

	logrus.Println("Starting server at", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
