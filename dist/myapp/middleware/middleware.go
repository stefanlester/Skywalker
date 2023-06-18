package middleware

import (
	"myapp/data"

	"github.com/stefanlester/skywalker"
)

type Middleware struct {
	App *skywalker.Skywalker
	Models data.Models
}