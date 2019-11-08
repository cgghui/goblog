package main

import "goblog/app"

func main() {
	app.New(app.Config{
		ConFile:    "../../config.ini",
		OpenLog:    true,
		ListenAddr: ":8787",
	})
}
