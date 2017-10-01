package main

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

type server struct {
	config   *config
	log      *logrus.Logger
	redis    *redis.Client
	template *template.Template
}

func newServer(c *config, l *logrus.Logger) (*server, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host + ":" + strconv.Itoa(c.Redis.Port),
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})
	if _, err := r.Ping().Result(); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %v", err)
	}

	t, err := template.ParseGlob("templates/*.xml")
	if err != nil {
		return nil, fmt.Errorf("unable to create template: %v", err)
	}

	return &server{
		config:   c,
		log:      l,
		redis:    r,
		template: t,
	}, nil
}
