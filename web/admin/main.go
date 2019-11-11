package main

import (
	"goblog/app"
	"goblog/web/admin/controller"
)

func main() {
	app.New("../../config.ini", []app.RouteBuilder{
		&controller.Common{},
	})
}
