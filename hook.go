package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

type Hook struct {
	Config  *Config
	From    string
	To      []string
	Options string
	Email   *email.Email
}

func (h *Hook) Do() {
	logrus.Debugf("hooking %s -> %s", h.From, h.To)
	h.Config.wg.Add(1)
	defer h.Config.wg.Done()
	err := h.send()
	if err != nil {
		// TODO: retry on error
		logrus.Errorf("error on sending hook: %s", err)
	}
}

func (h *Hook) send() error {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	w.WriteField("from", h.From)
	w.WriteField("to", fmt.Sprintf("[%v]", h.To))
	w.WriteField("options", h.Options)
	w.WriteField("subject", h.Email.Subject)
	w.WriteField("text", string(h.Email.Text))
	w.WriteField("html", string(h.Email.HTML))

	for _, at := range h.Email.Attachments {
		aw, err := w.CreateFormFile("attachment[]", at.Filename)
		if err != nil {
			logrus.Errorf("erro attach file from message: %s", err)
			io.Copy(aw, bytes.NewReader(at.Content))
		}
	}

	w.Close()

	r, _ := http.NewRequest("POST", h.Config.URI, body)
	r.Header.Add("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("http status code :%d", resp.StatusCode)
	}

	return nil
}
