package main

import (
	"context"
	"fmt"
	"github.com/cgghui/goblog/sys"
)

func main() {
	c := context.Background()
	r := sys.Redis.Conn(c)
	r.Close()
	fmt.Printf("%+v", r.Ping(c).String())
	fmt.Printf("%+v", sys.G)
}
