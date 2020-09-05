package main

import (
	"fmt"
	"log"
	"net/url"
)

func main() {
	target, err := url.Parse("redis://?db=1&auth=123123")
	if err != nil {
		log.Panicf("%+v\n", err)
	}
	fmt.Printf("%+v", target.Query().Get("ss"))
	//if err != nil {
	//	return nil
	//}
	//if target.Scheme != "redis" {
	//
	//}
	//if G.S.Addr == "" {
	//	G.S.Addr = ":6379"
	//} else {
	//	if strings.Index(G.S.Addr, ":") == -1 {
	//		G.S.Addr += ":6379"
	//	}
	//}
	//conn := redis.NewClient(&redis.Options{
	//	Network:  "tcp",
	//	Addr:     G.S.Addr,
	//	Password: G.S.Auth,
	//	DB:       int(G.S.DB),
	//})
}
