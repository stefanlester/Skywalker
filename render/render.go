package render

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
	// Debug disables the Go template cache so templates are re-parsed from
	// disk on every request, matching Jet's InDevelopmentMode behavior.
	Debug    bool
	JetViews *jet.Set
	Session  *scs.SessionManager

	templateCache   map[string]*template.Template
	templateCacheMu sync.RWMutex
}

type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
	Error           string
	Flash           string
}

func (s *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = s.Secure
	td.ServerName = s.ServerName
	td.CSRFToken = nosurf.Token(r)
	td.Port = s.Port
	if s.Session != nil {
		if s.Session.Exists(r.Context(), "userID") {
			td.IsAuthenticated = true
		}
		td.Error = s.Session.PopString(r.Context(), "error")
		td.Flash = s.Session.PopString(r.Context(), "flash")
	}
	return td
}

func (s *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(s.Renderer) {
	case "go":
		return s.GoPage(w, r, view, data)
	case "jet":
		return s.JetPage(w, r, view, variables, data)
	default:

	}
	return errors.New("no rendering engine specified")
}

// goTemplate returns the parsed Go template for view. Templates are parsed
// once and cached; when Debug is true the cache is bypassed and the template
// is re-parsed from disk on every call.
func (s *Render) goTemplate(view string) (*template.Template, error) {
	if s.Debug {
		return template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", s.RootPath, view))
	}

	s.templateCacheMu.RLock()
	tmpl, ok := s.templateCache[view]
	s.templateCacheMu.RUnlock()
	if ok {
		return tmpl, nil
	}

	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", s.RootPath, view))
	if err != nil {
		return nil, err
	}

	s.templateCacheMu.Lock()
	if s.templateCache == nil {
		s.templateCache = make(map[string]*template.Template)
	}
	s.templateCache[view] = tmpl
	s.templateCacheMu.Unlock()

	return tmpl, nil
}

// GoPage renders a standard Go template
func (s *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tmpl, err := s.goTemplate(view)
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	err = tmpl.Execute(w, &td)
	if err != nil {
		return err
	}

	return nil
}

// JetPage renders a template using the Jet templating engine
func (s *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = s.defaultData(td, r)

	t, err := s.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
