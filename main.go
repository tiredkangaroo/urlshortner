package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed version.txt
var version string

func getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func setRedirectionLinks(client *redis.Client, data *map[string]string) error {
	cmd := client.HGetAll(context.Background(), "urlshortner")
	if cmd.Err() != nil {
		return cmd.Err()
	}
	*data = cmd.Val()
	return nil
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// redirectionlinks provides a short -> long conversion.
	var redirectionlinks = make(map[string]string)

	fmt.Println(`Visit: http://[::1]:80`)

	go func() {
		for {
			err := setRedirectionLinks(client, &redirectionlinks)
			if err != nil {
				slog.Error("setting redirection linls", "error", err.Error())
			}
			time.Sleep(time.Second * 15)
		}
	}()

	err := startServer(&redirectionlinks)
	if err != nil {
		panic(err)
	}
}
