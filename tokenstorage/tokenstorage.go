package tokenstorage

/*
At the time of writing, there are 2 implementations that implement Storage:

	1. local - for local development. This stores tokens in a directory ./tokens
	2. ssm - for AWS deployment. This stores tokens in SSM Parameter store at /<SSM_PREFIX>/<email_at_emailhost>

Storage type is determined by envvar STORAGE. It defaults to ssm unless the value is 'local'.
*/

// Storage is an interface. Different types of token storage must implement it.
type Storage interface {
	StoreToken(string, interface{}) error
	RetrieveToken(string, interface{}) error
	ListEmails() ([]string, error)
}
