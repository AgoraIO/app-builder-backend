// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package utils

import (
	"io"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Taken from: https://gist.github.com/panta/2530672ca641d953ae452ecb5ef79d7d

// Config contains the configuration for logging
type Config struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
}

// Logger contains the logger object
type Logger struct {
	*zerolog.Logger
}

// SetLogLevel sets the level of the global logger
func SetLogLevel() {
	level := viper.GetString("LOG_LEVEL")
	var zerologLevel zerolog.Level
	if level == "PANIC" {
		zerologLevel = zerolog.PanicLevel
	} else if level == "FATAL" {
		zerologLevel = zerolog.FatalLevel
	} else if level == "ERROR" {
		zerologLevel = zerolog.ErrorLevel
	} else if level == "WARN" {
		zerologLevel = zerolog.WarnLevel
	} else if level == "INFO" {
		zerologLevel = zerolog.InfoLevel
	} else if level == "DEBUG" {
		zerologLevel = zerolog.DebugLevel
	} else {
		panic("Invalid Log Level")
	}

	zerolog.SetGlobalLevel(zerologLevel)
}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func Configure(config Config) *Logger {
	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}

	mw := io.MultiWriter(writers...)

	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(mw).With().Timestamp().Caller().Logger()
	SetLogLevel()

	logger.Info().
		Bool("fileLogging", config.FileLoggingEnabled).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config Config) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}
