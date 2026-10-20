package main

import (
	"code_base/internal/application"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	cmApp := application.NewCommerceApp(ctx)
	err := cmApp.Start(ctx)
	if err != nil {
		log.Fatalf("%s\n%v", "failed to start the application", err)
	}

	ListenForGracefulShutdown(cmApp)
}

func ListenForGracefulShutdown(application application.CommerceApp) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Application started")
	<-sigChan

	done := make(chan bool)

	log.Println("Shutting down...")
	go func() {
		err := application.Stop(context.Background())
		if err != nil {
			log.Println(err)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown complete")
	case <-time.After(time.Minute):
		log.Println("Shutdown timeout")
	}
}
