package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/yoramdelangen/iptv-web-app/internal/server"
	"github.com/yoramdelangen/iptv-web-app/internal/surreal"
	"github.com/yoramdelangen/iptv-web-app/internal/xtream"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("configuration")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.WatchConfig()

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
