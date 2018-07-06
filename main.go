package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Luzifer/rconfig"
	log "github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		FiFoFile       string `flag:"fifo" default:"/home/gameserver/terraria_cmd" description:"Path to create the fifo at"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

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

	cmd := exec.Command(rconfig.Args()[1], rconfig.Args()[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.WithError(err).Fatal("Unable to create stdin pipe")
	}

	// StdIN processing to send commands to Terraria
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go handleSignal(sigs, stdin)

	go inputLoop(stdin)

	// StdOUT processing to react on log output
	stdoutRead, stdoutWrite := io.Pipe()
	cmd.Stdout = stdoutWrite
	cmd.Stderr = stdoutWrite

	go outputLoop(stdoutRead, stdin)

	cmd.Run()
}

func outputLoop(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		log.Info(line)

		if strings.Contains(line, "has left.") {
			// A player left the server and after the last player
			// left the server will not save anymore so we'll do
			// that whenever a player leaves...
			fmt.Fprintln(out, "save")
		}
	}
}

func inputLoop(comm io.Writer) {
	for {
		f, err := os.Open(cfg.FiFoFile)
		if err != nil {
			log.WithError(err).Error("Unable to (re)open fifo, quitting now")
			fmt.Fprintln(comm, "exit")
			break
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fmt.Fprintln(comm, scanner.Text())
		}

		switch scanner.Err() {
		case nil:
			// This is fine
		case io.EOF:
			// This is fine
		default:
			log.WithError(err).Error("Unable to read from fifo")
			fmt.Fprintln(comm, "exit")
		}
	}
}

func handleSignal(sigs chan os.Signal, comm io.Writer) {
	sig := <-sigs
	log.WithField("signal", sig).Info("Received terminating singal")
	fmt.Fprintln(comm, "exit")
}
