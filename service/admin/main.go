package main

import (
	"goblog/app"
	"goblog/service/admin/controller"
)

func main() {
	app.New("../../config.ini", []app.RouteBuilder{
		&controller.Common{},
		&controller.Auth{},
	})
}
