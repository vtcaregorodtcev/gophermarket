package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/helpers"
)

func main() {
	fmt.Println("Starting the GopherMart Server...")

	addr := helpers.GetStringEnv("RUN_ADDRESS", flag.String("a", "", "server address"))
	dbURI := helpers.GetStringEnv("DATABASE_URI", flag.String("d", "", "db connection string"))

	flag.Parse()

	app := app.New(app.Config{Addr: *addr, DatabaseURI: *dbURI})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		fmt.Println("Received interrupt signal. Shutting down...")

		app.Shutdown()

		os.Exit(0)
	}()

	app.Run()
}
