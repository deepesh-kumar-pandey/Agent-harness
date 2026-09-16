package session

type Store interface {
	Get(id string) (*Session, error)
	Set(session *Session) error
	Delete(id string) error
}
