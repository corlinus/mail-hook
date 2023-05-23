package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type m struct{}

func resumeSpooler(cfg *Config) {
	files, _ := ioutil.ReadDir(cfg.SpoolDir)
	if len(files) == 0 {
		return
	}
	logrus.Infof("resume spooler. found %d files", len(files))
	w := make(chan m, cfg.SpoolThreads)

	for _, fi := range files {
		if cfg.ctx.Err() != nil {
			return
		}

		fname := fi.Name()
		if !strings.HasSuffix(fname, ".json") {
			continue
		}

		h, err := restore(filepath.Join(cfg.SpoolDir, fname), cfg)
		if err != nil {
			logrus.Errorf("can not read file %s: %s", fi.Name(), err)
			continue
		}

		w <- m{}
		go func() {
			h.Do(cfg.ctx, false)
			_ = <-w
		}()
	}
}
