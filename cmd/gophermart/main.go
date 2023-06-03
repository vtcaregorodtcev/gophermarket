package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/app"
)

func main() {
	fmt.Println("Starting the GopherMart Server...")

	app := app.New()

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
