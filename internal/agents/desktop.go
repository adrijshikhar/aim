package agents

// DesktopApp is an optional capability interface implemented by graphical desktop adapters.
type DesktopApp interface {
	IsDesktop() bool
	AppBundleName() string
	UserDataDir(profileDir string) string
}
