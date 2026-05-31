package main

import (
	"Velora/server/Internal/server"
	"Velora/server/Internal/server/clients"
	"Velora/server/Internal/server/config"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	port = flag.Int("port", 8080, "The port to listen on")
)

func main() {
	flag.Parse()

	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("Error loading config.env file")
	}

	var ctx = context.Background()

	var sheetService, err = sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	appConfig, err := config.NewAppConfig(ctx, sheetService, os.Getenv)
	if err != nil {
		log.Fatalf("Unable to load app config: %v", err)
	}

	var hub = server.NewHub(appConfig)

	http.HandleFunc("/velora", func(writer http.ResponseWriter, reader *http.Request) {
		hub.Serve(clients.NewWebsocketConnection, writer, reader)
	})

	go hub.Run()

	var addr = fmt.Sprintf(":%d", *port)

	err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Failed start server %v: ", err)
	}
}
