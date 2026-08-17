package agentfilestorage

/*
At the time of writing, there are 2 implementations that implement Storage:

	1. local - for local development. This stores tokens in at ./agentfiles/<email>/AGENT.md
	2. s3 - for AWS deployment. This stores tokens in SSM Parameter store at /<S3_BUCKET>/<email>/AGENT.md

Storage type is determined by envvar STORAGE. It defaults to ssm unless the value is 'local'.
*/

// Storage is an interface. Different types of storage must implement it.
type Storage interface {
	RetrieveFile(string) (string, error)
}

var (
	filename = "AGENT.md"
)
