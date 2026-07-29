package agentfilestorage

// Storage is an interface. Different types of storage must implement it.
type Storage interface {
	RetrieveFile(string) (string, error)
}

var (
	filename = "AGENT.md"
)
