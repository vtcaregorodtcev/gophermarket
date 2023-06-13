package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/vtcaregorodtcev/gophermarket/internal/app"
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
)

func main() {
	logger.NewLogger()
	logger.Infof("Starting the GopherMart Server...")

	addr := helpers.GetStringEnv("RUN_ADDRESS", flag.String("a", "", "server address"))
	accrualAddr := helpers.GetStringEnv("ACCRUAL_SYSTEM_ADDRESS", flag.String("r", "", "accrual service address"))
	dbURI := helpers.GetStringEnv("DATABASE_URI", flag.String("d", "", "db connection string"))

	flag.Parse()

	app := app.New(app.Config{Addr: *addr, DatabaseURI: *dbURI, AccrualAddr: *accrualAddr})
	defer app.Shutdown()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go app.Run()

	<-interrupt
	logger.Infof("Received interrupt signal. Shutting down...")
}
