package session

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestSession_InitSession(t *testing.T) {

	s := &Session{
		CookieLifetime: "100",
		CookiePersist:  "true",
		CookieName:     "skywalker",
		CookieDomain:   "localhost",
		SessionType:    "cookie",
	}

	var sm *scs.SessionManager

	ses := s.InitSession()

	var sessKind reflect.Kind
	var sessType reflect.Type

	rv := reflect.ValueOf(ses)

	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		fmt.Println("For loop", rv.Kind(), rv.Type(), rv)
		sessKind = rv.Kind()
		sessType = rv.Type()
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		t.Error("Session Manager is not valid, kind:", rv.Kind(), "type:", rv.Type())
	}

	if sessKind != reflect.ValueOf(sm).Kind() {
		t.Error("wrong kind retunred testing cookie session. Expected", reflect.ValueOf(sm).Kind(), "and got", sessKind)
	}

	if sessType != reflect.ValueOf(sm).Type() {
		t.Error("wrong kind retunred testing cookie session. Expected", reflect.ValueOf(sm).Type(), "and got", sessType)
	}
}
