package main

import (
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type logOpt struct {
	Logger   *log.Logger
	Color    *bool
	LogColor string
	Enabled  *bool
	File     *string
	Stdout   *bool
	Stderr   *bool
}

var cfgLoggingEnabled = config.Bool("logging.enabled", true)
var loggingOptions = map[string]logOpt{
	"trace": logOpt{
		Logger:   logger.Trace,
		Color:    config.Bool("logging.trace.color", true),
		LogColor: "!b",
		Enabled:  config.Bool("logging.trace.enabled", true),
		File:     config.String("logging.trace.file", ""),
		Stdout:   config.Bool("logging.trace.stdout", true),
		Stderr:   config.Bool("logging.trace.stderr", false),
	},
	"debug": logOpt{
		Logger:   logger.Debug,
		Color:    config.Bool("logging.debug.color", true),
		LogColor: "c",
		Enabled:  config.Bool("logging.debug.enabled", true),
		File:     config.String("logging.debug.file", ""),
		Stdout:   config.Bool("logging.debug.stdout", true),
		Stderr:   config.Bool("logging.debug.stderr", false),
	},
	"info": logOpt{
		Logger:   logger.Info,
		Color:    config.Bool("logging.info.color", true),
		LogColor: "g",
		Enabled:  config.Bool("logging.info.enabled", true),
		File:     config.String("logging.info.file", ""),
		Stdout:   config.Bool("logging.info.stdout", true),
		Stderr:   config.Bool("logging.info.stderr", false),
	},
	"warn": logOpt{
		Logger:   logger.Warn,
		Color:    config.Bool("logging.warn.color", true),
		LogColor: "y",
		Enabled:  config.Bool("logging.warn.enabled", true),
		File:     config.String("logging.warn.file", ""),
		Stdout:   config.Bool("logging.warn.stdout", true),
		Stderr:   config.Bool("logging.warn.stderr", false),
	},
	"error": logOpt{
		Logger:   logger.Error,
		Color:    config.Bool("logging.error.color", true),
		LogColor: "r",
		Enabled:  config.Bool("logging.error.enabled", true),
		File:     config.String("logging.error.file", ""),
		Stdout:   config.Bool("logging.error.stdout", true),
		Stderr:   config.Bool("logging.error.stderr", false),
	},
}

func applyLogSettings() (err error) {
	// Generate file pointers - you can send multiple levels to the same file
	files := make(map[string]*os.File, len(loggingOptions))
	for _, opts := range loggingOptions {
		_, exists := files[*opts.File]
		switch {
		case *opts.File == "":
			// None specified
			continue
		case exists:
			// Already exists
			continue
		default:
			if err = os.MkdirAll(filepath.Dir(*opts.File), 0755); err != nil {
				return
			}
			files[*opts.File], err = os.OpenFile(*opts.File, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				return
			}
		}
	}

	for _, opts := range loggingOptions {
		outs := make([]io.Writer, 0, 3)
		// Handle disable
		if !*cfgLoggingEnabled || !*opts.Enabled {
			outs = append(outs, ioutil.Discard)
			goto ApplyWriters
		}
		if *opts.File != "" {
			outs = append(outs, files[*opts.File])
		}
		if *opts.Stdout && *opts.Color {
			outs = append(outs, logger.NewColorStdout(opts.LogColor))
		}
		if *opts.Stdout && !*opts.Color {
			outs = append(outs, os.Stdout)
		}
		if *opts.Stderr && *opts.Color {
			outs = append(outs, logger.NewColorStderr(opts.LogColor))
		}
		if *opts.Stderr && !*opts.Color {
			outs = append(outs, os.Stderr)
		}

	ApplyWriters:
		var w io.Writer
		switch len(outs) {
		case 0:
			// noop
		case 1:
			w = outs[0]
		default:
			w = io.MultiWriter(outs...)
		}
		newLogger := log.New(w, opts.Logger.Prefix(), opts.Logger.Flags())
		*opts.Logger = *newLogger
	}

	return
}
