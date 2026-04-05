package version

// Set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
)

func String() string {
	if Version == "dev" {
		return "dev (" + Commit + ")"
	}
	return Version + " (" + Commit + ")"
}
