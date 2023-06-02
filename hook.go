package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

type Hook struct {
	Config  *Config `json:"-"`
	Fname   string  `json:"-"`
	From    string
	To      []string
	HookURI string
	Options string
	Email   *email.Email
}

func (h *Hook) String() string {
	return fmt.Sprintf("%s -> %s", h.From, h.To)
}

func (h *Hook) Do(ctx context.Context, dump bool) {
	logrus.Debugf("hooking %s", h)
	h.Config.wg.Add(1)
	defer h.Config.wg.Done()

	if dump {
		h.dump()
	}

	for i := 0; i < h.Config.SendReties; i++ {
		if i > 0 {
			logrus.Infof("retry hook: %s", h)
		}
		err := h.send()
		if err != nil {
			logrus.Errorf("error on sending hook: %s", err)

			sleep := time.Duration((i*i*i*i)+10) * time.Second
			t := time.NewTimer(sleep)

			select {
			case <-t.C:
				t.Stop()
				continue
			case <-ctx.Done():
				return
			}
		}

		h.removeFile()
		break
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
		_, err = io.Copy(aw, bytes.NewReader(at.Content))
		if err != nil {
			logrus.Errorf("can not copy email attachment to http request: %s", err)
		}
	}

	w.Close()
	return w.FormDataContentType(), body
}

func (h *Hook) send() error {
	logrus.Debugf("sending hook %s", h)
	contentType, body := h.body()
	r, _ := http.NewRequest("POST", h.HookURI, body)
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
		return fmt.Errorf("send error: %s", err)
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("send error: http status code :%d", resp.StatusCode)
	}

	logrus.Debugf("hook %s sent. http code %d", h, resp.StatusCode)
	return nil
}

func (h *Hook) dump() {
	h.createFname()
	f, err := os.Create(h.Fname)
	if err != nil {
		logrus.Errorf("can not create file %s: %s", h.Fname, err)
	}

	defer f.Close()
	err = json.NewEncoder(f).Encode(h)
	if err != nil {
		logrus.Errorf("can not dump file %s: %s", h.Fname, err)
	}
}

func (h *Hook) createFname() {
	hash := md5.New()
	io.WriteString(hash, h.From)
	io.WriteString(hash, fmt.Sprintf("%v", h.To))
	io.WriteString(hash, fmt.Sprintf("%d", time.Now().Unix()))
	h.Fname = filepath.Join(h.Config.SpoolDir, fmt.Sprintf("%x.json", hash.Sum(nil)))
}

func (h *Hook) removeFile() {
	err := os.Remove(h.Fname)
	if err != nil {
		logrus.Errorf("error delete file: %s", err)
	}
}

func restore(fname string, cfg *Config) (*Hook, error) {
	h := &Hook{Fname: fname, Config: cfg}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	err = json.NewDecoder(f).Decode(h)
	if err != nil {
		logrus.Errorf("can not restore file %s: %s", fname, err)
	}
	return h, nil
}
