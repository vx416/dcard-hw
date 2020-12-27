package logging

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// ColorfulLevelEncoder make level colorful
func ColorfulLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var (
		color int
		bold  bool
	)

	switch l {
	case zapcore.DebugLevel:
		color = colorBlue
	case zapcore.InfoLevel:
		color = colorGreen
	case zapcore.WarnLevel:
		color = colorYellow
	case zapcore.ErrorLevel:
		color = colorRed
	case zapcore.FatalLevel:
		color = colorRed
		bold = true
	case zapcore.PanicLevel:
		color = colorRed
		bold = true
	}

	s := colorize(l.CapitalString(), color)

	if bold {
		s = colorize(s, colorBold)
	}
	enc.AppendString(s)
}

// ColorizeCallerEncoder make caller colorful
func ColorizeCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorize(caller.TrimmedPath(), colorCyan))
}

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90
)

func colorize(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
