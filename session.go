package main

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/emersion/go-smtp"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

var Err = errors.New("internal server error")

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
	// TODO: blacklist
	s.To = append(s.To, to)
	logrus.Infof("mail %s -> %s", s.From, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	e, err := email.NewEmailFromReader(r)
	if err != nil {
		return err
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
	// logrus.Debugf("session reset")
	s.From = ""
	s.To = nil
	s.Options = smtp.MailOptions{}
}

func (s *Session) Logout() error {
	return nil
}
