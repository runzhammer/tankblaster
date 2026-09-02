package buildinfo

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func String() string {
	if Commit == "" || Commit == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (%s, %s)", Version, Commit, BuildTime)
}
