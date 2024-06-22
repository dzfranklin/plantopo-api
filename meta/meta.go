package meta

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

var Commit string
var CommitTime string
var Version string
var FullInfo string

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Warn("Failed to read build info")
		return
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			Commit = setting.Value
		} else if setting.Key == "vcs.time" {
			CommitTime = setting.Value
		}
	}

	if os.Getenv("APP_ENV") == "development" {
		Version = fmt.Sprintf("dev (%s)", time.Now().Format(time.RFC3339))
	} else {
		Version = fmt.Sprintf("%s (%s)", Commit, CommitTime)
	}

	FullInfo = fmt.Sprintf("plantopo-api %s\n\n%+v", Version, info)
}

func SetUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", fmt.Sprintf("github.com/dzfranklin/plantopo-api %s", Version))
}
