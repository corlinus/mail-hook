package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"

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

	c, b := h.body()
	err := h.send(c, b)
	if err != nil {
		// TODO: retry on error
		logrus.Errorf("error on sending hook: %s", err)
	}
}

func (h *Hook) body() (string, *bytes.Buffer) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	w.WriteField("smtp_from", h.From)
	w.WriteField("smtp_to", fmt.Sprintf("%v", h.To))
	w.WriteField("smpt_options", h.Options)
	w.WriteField("from", h.Email.From)
	w.WriteField("to", fmt.Sprintf("%v", h.Email.To))
	w.WriteField("subject", h.Email.Subject)
	w.WriteField("text", string(h.Email.Text))
	w.WriteField("html", string(h.Email.HTML))

	for i, at := range h.Email.Attachments {
		fname := at.Filename
		if len(fname) == 0 {
			fname = fmt.Sprintf("filename%2d", i+1)
		}
		aw, err := w.CreateFormFile("attachment[]", fname)
		if err != nil {
			logrus.Errorf("error attach file from message: %s", err)
		}
		i, err := io.Copy(aw, bytes.NewReader(at.Content))
		if err != nil {
			logrus.Errorf("can not copy email attachment to http request: ", err)
		}

		logrus.Debugf("bytes copied %d", i)
	}

	w.Close()
	return w.FormDataContentType(), body
}

func (h *Hook) send(contentType string, body io.Reader) error {
	logrus.Debugf("sending hook for %s -> %v", h.From, h.To)

	r, _ := http.NewRequest("POST", h.Config.URI, body)
	r.Header.Add("Content-Type", contentType)

	if logrus.GetLevel() >= logrus.TraceLevel {
		reqDump, err := httputil.DumpRequestOut(r, true)
		if err != nil {
			logrus.Errorf("err : %s", err)
		}
		fmt.Printf("REQUEST:\n%s", string(reqDump))
	}

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("http status code :%d", resp.StatusCode)
	}

	logrus.Debugf("hook %s -> %v sent. http code %d", h.From, h.To, resp.StatusCode)
	return nil
}
