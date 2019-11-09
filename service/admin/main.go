package main

import "goblog/app"

func main() {
	app.New(app.Config{
		ConfigFilePath: "config.ini",
	})
}
