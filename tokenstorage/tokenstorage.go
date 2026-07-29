package tokenstorage

// Storage is an interface. Different types of token storage must implement it.
type Storage interface {
	StoreToken(string, interface{}) error
	RetrieveToken(string, interface{}) error
	ListEmails() ([]string, error)
}
