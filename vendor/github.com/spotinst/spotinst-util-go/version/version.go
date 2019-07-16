package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// These variables are initialized via the linker -X flag in the
// top-level Makefile when compiling release binaries.
var (
	// The git commit that was compiled. These will be filled in by the
	// compiler.
	GitCommit   string
	GitDescribe string

	// The main version number that is being run at the moment.
	Version = "unknown"

	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a
	// pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease string

	// Build time in ISO-8601 format.
	BuildDate = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Build platform.
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

type Info struct {
	GoVersion         string
	Platform          string
	BuildDate         string
	GitCommit         string
	GitDescribe       string
	Version           string
	VersionPrerelease string
}

// GetInfo composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetInfo() Info {
	info := Info{
		GoVersion:         runtime.Version(),
		Platform:          Platform,
		BuildDate:         BuildDate,
		GitCommit:         GitCommit,
		GitDescribe:       GitDescribe,
		Version:           Version,
		VersionPrerelease: VersionPrerelease,
	}
	return info
}

func (i Info) String() string {
	version := i.Version
	if i.GitDescribe != "" && !strings.HasPrefix(i.GitCommit, i.GitDescribe) {
		version = i.GitDescribe
	}

	release := i.VersionPrerelease
	if release != "" {
		version += fmt.Sprintf("-%s", release)
		if i.GitCommit != "" {
			version += fmt.Sprintf(", build %s", i.GitCommit)
		}
	}

	// Strip off any single quotes added by the git information.
	version = strings.Replace(version, "'", "", -1)

	return fmt.Sprintf("%s (%s, built %s, %s)",
		version, i.Platform, i.BuildDate, i.GoVersion)
}
