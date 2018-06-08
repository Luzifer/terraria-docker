package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luzifer/rconfig"
	log "github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		FiFoFile       string `flag:"fifo" default:"/home/gameserver/terraria_cmd" description:"Path to create the fifo at"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	done = make(chan struct{}, 1)

	version = "dev"
)

func init() {
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("terraria-docker %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	if err := syscall.Mkfifo(cfg.FiFoFile, 0644); err != nil {
		log.WithError(err).Fatal("Unable to create fifo")
	}
	defer os.Remove(cfg.FiFoFile)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go handleSignal(sigs)

	go inputLoop()

	for {
		select {
		case <-done:
			log.Info("Waiting 5s until quit")
			<-time.After(5 * time.Second) // Allow some time for terraria to quit gracefully
			return
		}
	}
}

func inputLoop() {
	for {
		f, err := os.Open(cfg.FiFoFile)
		if err != nil {
			log.WithError(err).Error("Unable to (re)open fifo, quitting now")
			done <- struct{}{}
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			if scanner.Text() == "exit" {
				done <- struct{}{}
			}
		}

		switch scanner.Err() {
		case nil:
			// This is fine
		case io.EOF:
			// This is fine
		default:
			log.WithError(err).Error("Unable to read from fifo")
			fmt.Println("exit")
			done <- struct{}{}
		}
	}
}

func handleSignal(sigs chan os.Signal) {
	sig := <-sigs
	log.WithField("signal", sig).Info("Received terminating singal")
	fmt.Println("exit")
	done <- struct{}{}
}
