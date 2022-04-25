package logconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

const logLevelKey = "GAZELLE_LANGUAGES_JAVA_LOG_LEVEL"

var strToJava = map[string]string{
	"trace":    "trace",
	"debug":    "debug",
	"info":     "info",
	"warn":     "warn",
	"error":    "error",
	"fatal":    "error",
	"panic":    "error",
	"off":      "off",
	"disabled": "off",
}

func LogLevel() (zerolog.Level, string) {
	v := os.Getenv(logLevelKey)
	if v == "" {
		v = "info"
	}

	goLevel, err := zerolog.ParseLevel(strings.ToLower(v))
	if err != nil {
		panic(fmt.Sprintf("invalid go level '%s': %s", v, err))
	}

	javaLevel, found := strToJava[strings.ToLower(v)]
	if !found {
		panic("invalid go level: " + v)
	}

	return goLevel, javaLevel
}
