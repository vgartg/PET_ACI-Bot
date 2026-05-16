package executor

type Executor interface {
	OpenFile(path string) error
	OpenURL(rawURL string) error
	KillProcess(name string) error
	Shutdown() error
}
