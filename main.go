package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Join           []string      `flag:"join,j" default:"" description:"Channels to join (specify multiple times for multiple channels)"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		Nick           string        `flag:"nick" default:"" description:"Nick to choose after connecting (defaults to user)"`
		Port           int64         `flag:"port" default:"6667" description:"Port to connect to"`
		Quiet          bool          `flag:"quiet,q" default:"false" description:"Do not print messages to stdout"`
		SendBurst      int           `flag:"send-burst" default:"0" description:"Number of messages to be sent in a burst"`
		SendLimit      time.Duration `flag:"send-limit" default:"0" description:"How long to wait between two messages"`
		Server         string        `flag:"server,s" default:"" description:"IRC Server to connect to"`
		ServerPass     string        `flag:"server-pass,p" default:"" description:"Password to authenticate"`
		TLS            bool          `flag:"tls" default:"false" description:"Use TLS connection"`
		User           string        `flag:"user,u" default:"tinyirc" description:"User to use to connect to the server"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	connWaitOnce    = new(sync.Once)
	connEstablished = new(sync.WaitGroup)
	done            bool

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("tinyirc %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	connEstablished.Add(1)

	client, conn, err := connect()
	if err != nil {
		log.WithError(err).Fatal("Unable to connect")
	}
	defer conn.Close()

	go func() {
		if err := client.Run(); err != nil && !done {
			log.WithError(err).Fatal("IRC client reported error")
		}
	}()

	connEstablished.Wait()

	defer client.WriteMessage(&irc.Message{Command: "QUIT"})

	for _, c := range cfg.Join {
		logger := log.WithField("channel", c)
		logger.Debug("Joining channel")
		if err = client.WriteMessage(&irc.Message{
			Command: "JOIN",
			Params:  []string{c},
		}); err != nil {
			logger.WithError(err).Error("Unable to join channel")
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		logger := log.WithField("line", line)

		logger.Debug("Processing line")
		msg, err := irc.ParseMessage(line)
		if err != nil {
			logger.WithError(err).Error("Unable to parse line")
			continue
		}

		if err = client.WriteMessage(msg); err != nil {
			logger.WithError(err).Error("Unable to send message")
		}
	}

	done = true
}

func connect() (*irc.Client, net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)

	for f, r := range map[string]bool{
		"server": cfg.Server != "",
		"user":   cfg.User != "",
	} {
		if !r {
			return nil, nil, errors.Errorf("missing configuration: %s", f)
		}
	}

	if cfg.TLS {
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Server, cfg.Port), nil)
	} else {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Server, cfg.Port))
	}

	if err != nil {
		return nil, nil, errors.Wrap(err, "creating tcp connection")
	}

	nick := cfg.Nick
	if nick == "" {
		nick = cfg.User
	}

	return irc.NewClient(conn, irc.ClientConfig{
		Nick: nick,
		Pass: cfg.ServerPass,
		User: cfg.User,

		SendBurst: cfg.SendBurst,
		SendLimit: cfg.SendLimit,

		Handler: irc.HandlerFunc(printMessage),
	}), conn, nil
}

func printMessage(c *irc.Client, m *irc.Message) {
	if m.Command == "001" {
		connWaitOnce.Do(connEstablished.Done)
	}

	if _, err := strconv.Atoi(m.Command); err == nil {
		// Numeric command, connection setup, do not print
		return
	}

	if cfg.Quiet {
		return
	}

	fmt.Println(strings.TrimSpace(m.String()))
}
