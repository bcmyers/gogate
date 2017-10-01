package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sebest/logrusly"
)

func main() {
	time.Sleep(time.Second * 5)

	log := logrus.New()

	c := newConfig()
	if err := c.loadFromJSON("config.json"); err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	hook := logrusly.NewLogglyHook(c.APIKeyLoggly, "https://logs-01.loggly.com/bulk/", logrus.InfoLevel, "gogate")
	log.Hooks.Add(hook)
	go flushLog(hook)

	server, err := newServer(c, log)
	if err != nil {
		log.Fatalf("unable to initialize server: %v", err)
	}

	hostname := c.Host + ":" + strconv.Itoa(c.Port)

	http.HandleFunc("/", server.handler)
	fmt.Printf("Listening on %v\n", hostname)
	log.Fatal(http.ListenAndServe(hostname, nil))
}

func flushLog(hook *logrusly.LogglyHook) {
	for {
		time.Sleep(time.Second * 60 * 10)
		hook.Flush()
	}
}
