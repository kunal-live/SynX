package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"synx/internal/app"
)

//go:embed web/*
var webFS embed.FS

func main() {
	dirFlag := flag.String("dir", "", "directory shared by SynX")
	portFlag := flag.Int("port", 0, "HTTP port (default from config: 8787)")
	flag.Parse()

	synxApp, err := app.New(webFS)
	if err != nil {
		log.Fatalf("Failed to initialize SynX: %v", err)
	}

	if *dirFlag != "" {
		_ = synxApp.SetSharedDir(*dirFlag)
	}
	if *portFlag != 0 {
		synxApp.Config.Network.Port = *portFlag
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	synxApp.Startup(ctx)

	state := synxApp.GetState()
	log.Printf("==================================================")
	log.Printf(" SynX Core Daemon Running")
	log.Printf(" Device ID:   %s", synxApp.Identity.DeviceID)
	log.Printf(" Device Name: %s", synxApp.Identity.DeviceName)
	log.Printf(" Address:     http://%s", state["address"])
	log.Printf(" Shared Dir:  %s", state["sharedDir"])
	log.Printf(" Pair Token:  %s", synxApp.Token)
	log.Printf("==================================================")

	<-ctx.Done()
	log.Println("Shutting down SynX...")
	synxApp.Shutdown(ctx)
}
