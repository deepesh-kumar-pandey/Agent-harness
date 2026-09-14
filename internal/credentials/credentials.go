package credentials

type CredentialStore interface {
	Set(provider string, key string) error
	Get(provider string) (string, error)
	Delete(provider string) error
}
