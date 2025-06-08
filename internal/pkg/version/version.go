package version

import "fmt"

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	Author    = "Kimi Tsai"
	Email     = "kimi0230@gmail.com"
)

// GetVersion 返回完整的版本資訊
func GetVersion() string {
	return fmt.Sprintf("Version: %s\nBuild Time: %s\nGit Commit: %s\nAuthor: %s <%s>",
		Version, BuildTime, GitCommit, Author, Email)
}

// GetShortVersion 返回簡短版本資訊
func GetShortVersion() string {
	return Version
}
