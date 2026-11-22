package appinfo

import (
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/vnafikov/common/datetime"

	"github.com/vnafikov/foundation/logger"
)

const goVersionPrefix = "go"

type templateData struct {
	Options
	LogLevel       string
	GoVersion      string
	TimeZoneName   string
	TimeZoneOffset string
	CPUs           int
	PID            int
	PPID           int
}

func newTemplateData(options Options) templateData {
	timeZoneName, timeZoneOffset := time.Now().Zone()
	return templateData{
		Options:        options,
		LogLevel:       strings.ToLower(logger.Level.String()),
		GoVersion:      goVersion(),
		TimeZoneName:   timeZoneName,
		TimeZoneOffset: datetime.FormatTimeZoneOffset(timeZoneOffset),
		CPUs:           runtime.NumCPU(),
		PID:            os.Getpid(),
		PPID:           os.Getppid(),
	}
}

func goVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		ver, _ := strings.CutPrefix(info.GoVersion, goVersionPrefix)
		return ver
	}
	return ""
}
