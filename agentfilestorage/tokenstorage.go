package agentfilestorage

type Storage interface {
	RetrieveFile(string) (string, error)
}
