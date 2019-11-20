package main

import (
	"goblog/app"
	"goblog/service/admin/controller"
)

func main() {
	app.New([]app.RouteBuilder{
		&controller.Common{},
		&controller.Auth{},
	})
}
