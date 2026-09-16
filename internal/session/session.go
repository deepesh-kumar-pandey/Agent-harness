package session

type Session struct {
	ID string
}

func NewSession(id string) *Session {
	return &Session{
		ID: id,
	}
}
