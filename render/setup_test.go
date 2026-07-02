package render

import (
	"os"
	"testing"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./testdata/views"),
	jet.InDevelopmentMode(),
)

var testSession = scs.New()

var testRenderer = Render{
	Renderer: "",
	RootPath: "",
	JetViews: views,
}

func TestMain(m *testing.M) {
	testSession.Lifetime = 24 * time.Hour
	testRenderer.Session = testSession

	os.Exit(m.Run())
}
