package logging

import (
	"errors"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// ErrUnknownLoggerType represent unknown logger type  error
	ErrUnknownLoggerType = errors.New("logger type unknown")
)

// LoggerType define logger engine type
type LoggerType string

const (
	// Zap represent zap logger engine
	Zap LoggerType = "zap"
)

// Env define logger environment
type Env string

// IsDev make env to lower and check if string is dev
func (env Env) IsDev() bool {
	envStr := strings.ToLower(string(env))

	return envStr == "dev" || envStr == "development"
}

// Level define logger level
type Level string

// ZapLevel convert string to zapcore.Level
func (l Level) ZapLevel() zapcore.Level {
	var (
		zapLevel zapcore.Level
		level    = strings.ToLower(string(l))
	)
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	case "panic":
		zapLevel = zapcore.PanicLevel
	}
	return zapLevel
}

func New(c *Config) (Logger, error) {
	return c.Build()
}

// Config represetn logger config
type Config struct {
	AppName        string   `yaml:"app_name" env:"APPNAME"`
	Env            Env      `yaml:"env" env:"ENV"`
	Level          Level    `yaml:"level" env:"LEVEL"`
	OutputPaths    []string `yaml:"output_paths" env:"OUTPUTS"`
	ErrOutputPaths []string `yaml:"err_output_paths" env:"ERR_OUTPUTS"`
}

// Build build logger
func (config Config) Build() (Logger, error) {
	var (
		logger Logger
		err    error
	)

	logger, err = config.buildZap()
	if err != nil {
		return nil, err
	}

	SetGlobal(logger)
	return logger, nil
}

func (config Config) buildZap() (Logger, error) {
	var (
		zapConfig zap.Config
	)
	if config.Env.IsDev() {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(config.Level.ZapLevel())
		zapConfig.EncoderConfig.EncodeLevel = ColorfulLevelEncoder
		zapConfig.EncoderConfig.EncodeCaller = ColorizeCallerEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, config.OutputPaths...)
	}
	if len(config.ErrOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = append(zapConfig.OutputPaths, config.ErrOutputPaths...)
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	fields := zap.Fields(
		zap.String("app_name", config.AppName),
		zap.String("host", host),
	)

	zaplog, err := zapConfig.Build(fields, zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &ZapAdapter{zaplog: zaplog, env: config.Env}, nil
}
