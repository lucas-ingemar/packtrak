package shared

type (
	CommandName string
	ManagerName string
)

const (
	CommandInstall CommandName = "install"
	CommandRemove  CommandName = "remove"
	CommandList    CommandName = "list"
	CommandSync    CommandName = "sync"
)

type Package struct {
	Name          string
	FullName      string
	Version       string
	LatestVersion string
	RepoUrl       string
}

type Dependency struct {
	Name     string
	FullName string
	// Version       string
	// LatestVersion string
	// RepoUrl       string
}
