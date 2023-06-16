package skywalker

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)


func (s *Skywalker) SessionLoad(next http.Handler) http.Handler {
	s.InfoLog.Println("SessionLoad Called")
	return s.Session.LoadAndSave(next)
}

func (s *Skywalker) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(s.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: secure,
		SameSite: http.SameSiteStrictMode,
		Domain: c.config.cookie.domain,
	})

	return csrfHandler
}
