package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"

	"github.com/emersion/go-smtp"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

var ErrInternal = errors.New("internal server error")

type Session struct {
	Config  *Config
	From    string
	To      []string
	Options smtp.MailOptions
}

func (s *Session) AuthPlain(username, password string) error {
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	if s.Config == nil {
		return errors.New("no session config")
	}
	s.From = from
	s.Options = *opts
	return nil
}

func (s *Session) Rcpt(to string) error {
	var allow bool
	a, err := mail.ParseAddress(to)
	if err != nil {
		return err
	}

	addr := a.Address
	at := strings.LastIndex(addr, "@")
	if at == -1 {
		return errors.New("invalid email address")
	}
	rcptDomain := addr[at+1:]
	for _, domain := range s.Config.AllowDomains {
		if rcptDomain == domain {
			allow = true
			break
		}
	}
	if !allow {
		return fmt.Errorf("not allowed %s", to)
	}

	s.To = append(s.To, to)
	logrus.Infof("mail %s -> %s", s.From, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	e, err := email.NewEmailFromReader(r)
	if err != nil {
		return err
	}

	if logrus.GetLevel() >= logrus.TraceLevel {
		bytes, _ := e.Bytes()
		fmt.Printf("EMAIL:\n%s", string(bytes))
	}

	opts, _ := json.Marshal(s.Options)
	hook := &Hook{
		Config:  s.Config,
		Email:   e,
		From:    s.From,
		To:      s.To,
		Options: string(opts),
	}
	go hook.Do()

	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
	s.Options = smtp.MailOptions{}
}

func (s *Session) Logout() error {
	return nil
}
