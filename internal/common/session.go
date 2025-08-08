package common

// Session struct for session/cookie management, authentication, and CAPTCHA handling.
type Session struct {
	Cookies   map[string]string
	AuthToken string
}

func NewSession() *Session {
	return &Session{Cookies: make(map[string]string)}
}
