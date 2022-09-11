package skywalker

import "net/http"

func (s *Skywalker) SessionLoad(next http.Handler) http.Handler {
	return s.Session.LoadAndSave(next)
}
