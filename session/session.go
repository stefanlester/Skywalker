package session

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
)

type Session struct {
	CookieLifetime string
	CookieName     string
	CookieDomain   string
	CookiePersist  string
	SessionType    string
	CookieSecure   string
}

func (s *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should the cookie last?
	minutes, err := strconv.Atoi(s.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// should the cookie persist?
	if strings.ToLower(s.CookiePersist) == "true" {
		persist = true
	}

	// should the cookie be secure?
	if strings.ToLower(s.CookieSecure) == "https" {
		secure = true
	} else {
		secure = false
	}

	// create a new session manager
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = s.CookieDomain
	session.Cookie.Secure = secure
	session.Cookie.Domain = s.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store
	switch strings.ToLower(s.SessionType) {
	case "redis":

	case "mysql", "mariadb":

	case "postgres", "postgresql":

	default: // default to in-memory
	}

	return session
}
