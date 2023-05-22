package main

import (
	"context"
	"time"

	"github.com/emersion/go-smtp"
)

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

// The Backend implements SMTP server methods.
type Backend struct {
	config *Config
}

func (b *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{Config: b.config}, nil
}

func newServer(cfg *Config) *smtp.Server {
	be := &Backend{cfg}
	s := smtp.NewServer(be)

	s.Addr = cfg.Addr
	s.Domain = cfg.Domain
	s.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	s.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	s.MaxMessageBytes = cfg.MaxMessageBytes
	s.MaxRecipients = cfg.MaxRecipients

	s.AuthDisabled = true

	return s
}
