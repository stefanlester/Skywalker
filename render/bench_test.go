package render

import (
	"net/http"
	"testing"
)

// discardWriter is a minimal http.ResponseWriter that throws away the body,
// keeping the benchmark focused on template lookup, parse, and execute cost.
type discardWriter struct{ header http.Header }

func (d *discardWriter) Header() http.Header         { return d.header }
func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }
func (d *discardWriter) WriteHeader(statusCode int)  {}

// BenchmarkGoPage measures the steady-state cost of rendering a Go template
// via GoPage, using the fixtures in testdata. It uses its own Render instance
// so mutations made by other tests to the shared testRenderer cannot skew it.
func BenchmarkGoPage(b *testing.B) {
	renderer := &Render{
		Renderer: "go",
		RootPath: "./testdata",
	}

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	w := &discardWriter{header: make(http.Header)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := renderer.GoPage(w, r, "home", nil); err != nil {
			b.Fatal(err)
		}
	}
}
