package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"os"

	"github.com/stefanlester/skywalker"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// init skywalker
	skywalker := &skywalker.Skywalker{}
	err = skywalker.New(path)
	if err != nil {
		log.Fatal(err)
	}

	skywalker.AppName = "myapp"

	// init handlers
	myHandlers := &handlers.Handlers{
		App: skywalker,
	}

	app := &application{
		App:      skywalker,
		Handlers: myHandlers,
	}

	app.App.Routes = app.routes()

	app.Models = data.New(app.App.DB.Pool)

	return app
}
