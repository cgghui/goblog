package main

import (
	"goblog/app"
	"goblog/web/admin/controller"
)

func main() {
	app.New([]app.RouteBuilder{
		&controller.Common{},
	})
}
