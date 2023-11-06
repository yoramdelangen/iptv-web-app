package main

import (
	"fmt"
	"log"

	"github.com/yoramdelangen/iptv-web-app/internal/server"
	"github.com/yoramdelangen/iptv-web-app/internal/surreal"
	"github.com/yoramdelangen/iptv-web-app/internal/xtream"
)

func main() {
	// SyncXtreamApi()
	Server()
}

func Server() {
	app := server.New()

	log.Fatal(app.Listen("localhost:3000"))
}

func SyncXtreamApi() {
	api := xtream.New(surreal.DB)

	api.RunAll()
	api.CategoryStats()

	fmt.Printf("API XTREAM %+v\n", api)
}
