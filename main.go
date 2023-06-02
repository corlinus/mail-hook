package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
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

			go func() {
				time.Sleep(20 * time.Second)
				cfg.cancel()
			}()

			err := s.Shutdown(cfg.ctx)
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
	prepare(cfg)

	cfg.ctx, cfg.cancel = context.WithCancel(context.Background())

	s := newServer(cfg)
	sygnals(s, cfg)
	health()

	resumeSpooler(cfg)

	logrus.Println("Starting server at", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}

func prepare(cfg *Config) {
	logrus.SetLevel(cfg.LogLevel)
	if cfg.SpoolThreads < 1 {
		cfg.SpoolThreads = 1
	}

	err := os.MkdirAll(cfg.SpoolDir, os.ModePerm)
	if err != nil {
		logrus.Fatalf("can not create spool dir %s: %s", cfg.SpoolDir, err)
	}
}

func health() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	go http.ListenAndServe(":8080", nil)
}
