package skywalker

import "net/http"

func (s *Skywalker) SessionLoad(next http.Handler) http.Handler {
	s.InfoLog.Println("SessionLoad Called")
	return s.Session.LoadAndSave(next)
}
