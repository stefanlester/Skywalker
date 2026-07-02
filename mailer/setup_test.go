package mailer

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)


var pool *dockertest.Pool
var resource *dockertest.Resource
var dockerAvailable bool

var mailer = Mail{
	Domain: "localhost",
	Templates: "./testdata/mail",
	Host: "localhost",
	Port: 1026,
	Encryption: "none",
	FromAddress: "me@here.com",
	FromName: "Joe",
	Jobs: make(chan Message, 1),
	Results: make(chan Result, 1),
}

func TestMain(m *testing.M) {
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Println("could not connect to docker — skipping mailer container tests:", err)
		go mailer.ListenForMail()
		os.Exit(m.Run())
	}
	pool = p

	// If the Docker daemon is not reachable, run the tests that do not need a
	// container and skip the ones that do (they guard on dockerAvailable).
	if err := pool.Client.Ping(); err != nil {
		log.Println("Docker not available — skipping mailer container tests")
		go mailer.ListenForMail()
		os.Exit(m.Run())
	}
	dockerAvailable = true

	opts := dockertest.RunOptions{
		Repository: "mailhog/mailhog",
		Tag: "latest",
		Env: []string{},
		ExposedPorts: []string{"1025", "8025"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"1025": {
				{HostIP: "0.0.0.0", HostPort: "1026"},
			},
			"8025": {
				{HostIP: "0.0.0.0", HostPort: "8026"},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		log.Println(err)
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatal("Could not start resource")
	}

	time.Sleep(2 * time.Second)

	go mailer.ListenForMail()

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}