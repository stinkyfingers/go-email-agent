package tokenstorage

type Storage interface {
	StoreToken(string, interface{}) error
	RetrieveToken(string, interface{}) error
	ListEmails() ([]string, error)
}
