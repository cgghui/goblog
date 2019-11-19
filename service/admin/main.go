package main

import (
	"fmt"
	"goblog/app"
	"goblog/model"
	"goblog/service/admin/controller"
	"time"
)

func main() {
	t := time.Now()
	pwd := model.AdminGeneratePassword("123123")
	ret := model.AdminVerifyPassword("$2a$10$X0VV5YWrmowEpiqnVxPk0e8VFQBmwWKrk.AIeFWUOgY8uPrS2iFcO", "123123123")
	fmt.Printf("%v", ret)
	elapsed := time.Since(t)
	fmt.Printf("app elapsed: %v result: %s", elapsed, pwd)
	app.New([]app.RouteBuilder{
		&controller.Common{},
		&controller.Auth{},
	})
}
