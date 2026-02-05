package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/stieneee/gumble/gumble"
	"github.com/stieneee/gumble/gumbleopenal"
	"github.com/stieneee/gumble/gumbleutil"
	_ "github.com/stieneee/gumble/opus"
)

type Config struct {
	MumbleAddress  string
	MumbleUsername string
	MumblePassword string
	MumbleChannel  string
}

func LoadConfig() Config {
	ma, maP := os.LookupEnv("MUMBLE_ADDRESS")
	if !maP {
		log.Fatal("MUMBLE_ADDRESS env var must be set")
	}

	mu, muP := os.LookupEnv("MUMBLE_USERNAME")
	if !muP {
		hostname, err := os.Hostname()
		if err != nil {
			panic("os.Hostname() errored: " + err.Error())
		}

		slog.Warn("MUMBLE_USERNAME is empty, using default BiB_{hostname} username")
		mu = fmt.Sprintf("BiB_%s", hostname)
	}

	mp, mpP := os.LookupEnv("MUMBLE_PASSWORD")
	if !mpP {
		slog.Warn("MUMBLE_PASSWORD is empty, using default empty password")
		mp = ""
	}

	mc, mcP := os.LookupEnv("MUMBLE_CHANNEL")
	if !mcP {
		slog.Warn("MUMBLE_CHANNEL is empty, using default root channel")
		mc = ""
	}

	return Config{
		MumbleAddress:  ma,
		MumbleUsername: mu,
		MumblePassword: mp,
		MumbleChannel:  mc,
	}
}

func Connect(c *gumble.Config, addr string) {
	client, err := GumbleDialInsecure(addr, c)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Connected to the Mumble server", "address", addr, "username", c.Username)

	if os.Getenv("ALSOFT_LOGLEVEL") == "" {
		os.Setenv("ALSOFT_LOGLEVEL", "0")
	}
	if _, err := gumbleopenal.New(client); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	select {}
}

// Same as gumble.Dial() but with InsecureSkipVerify set to true
func GumbleDialInsecure(addr string, config *gumble.Config) (*gumble.Client, error) {
	return gumble.DialWithDialer(new(net.Dialer), addr, config, &tls.Config{
		InsecureSkipVerify: true,
	})
}

func main() {
	c := LoadConfig()
	gc := gumble.NewConfig()
	gc.Username = c.MumbleUsername
	gc.Password = c.MumblePassword
	gc.Attach(gumbleutil.AutoBitrate)
	gc.Attach(&gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			if c.MumbleChannel == "" {
				slog.Info("Joined default channel")
				return
			}

			ch := e.Client.Channels.Find(c.MumbleChannel)
			if ch == nil {
				slog.Warn("Channel not found", "name", c.MumbleChannel)
				return
			}
			e.Client.Self.Move(ch)
			slog.Info("Joined channel", "name", ch.Name)
		},
	})
	Connect(gc, c.MumbleAddress)
}
