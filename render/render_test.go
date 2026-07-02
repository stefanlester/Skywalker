package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// withSession loads a session into the request context so the renderer's
// defaultData can read/write session values without a nil-pointer panic.
func withSession(r *http.Request) *http.Request {
	ctx, _ := testSession.Load(r.Context(), "")
	return r.WithContext(ctx)
}

var pageData = []struct {
	name          string
	renderer      string
	template      string
	errorExpected bool
	errorMessage  string
}{
	{"go_page", "go", "home", false, "error rendering go template"},
	{"go_page_no_template", "go", "no-file", true, "no error rendering non-existent go template, when one is expected"},
	{"go_page", "jet", "home", false, "error rendering jet template"},
	{"go_page_no_template", "jet", "no-file", true, "no error rendering non-existent jet template, when one is expected"},
	{"invalid_render_engine", "foo", "home", true, "no error rendering with invalid rendering engine"},
}

func TestRender_Page(t *testing.T) {

	for _, e := range pageData {
		r, err := http.NewRequest("GET", "/some-url", nil)
		if err != nil {
			t.Fatal(err)
		}

		r = withSession(r)

		w := httptest.NewRecorder()

		testRenderer.Renderer = e.renderer
		testRenderer.RootPath = "./testdata"

		err = testRenderer.Page(w, r, e.template, nil, nil)
		if e.errorExpected {
			if err == nil {
				t.Errorf("%s: %s: expected error, but none was returned", e.name, e.errorMessage)
			}
		} else {
			if err != nil {
				t.Errorf("%s: %s: %s: expected error, but none was returned", e.name, e.errorMessage, err.Error())
			}
		}
	}
}

func TestRender_GoPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Fatal(err)
	}

	r = withSession(r)

	testRenderer.Renderer = "go"
	testRenderer.RootPath = "./testdata"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRender_JetPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/url", nil)
	if err != nil {
		t.Fatal(err)
	}

	r = withSession(r)

	testRenderer.Renderer = "jet"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}
