package main

import (
	"log"
	"net/http"

	"github.com/stefanlester/skywalker/filesystems/miniofilesystem"

	"github.com/stefanlester/skywalker"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes here
	a.get("/", a.Handlers.Home)
	a.get("/test-minio", func(w http.ResponseWriter, r *http.Request) {
		f := a.App.FileSystems["MINIO"].(miniofilesystem.Minio)

		files, err := f.List("")
		if err != nil {
			log.Println(err)
			return
		}

		for _, file := range files {
			log.Println(file.Key)
		}
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	// routes from celeritas
	a.App.Routes.Mount("/skywalker", skywalker.Routes())
	a.App.Routes.Mount("/api", a.ApiRoutes())

	return a.App.Routes
}
