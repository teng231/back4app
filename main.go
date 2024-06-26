package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DbDSN         string `env:"DB_DSN"`
	Port          string `env:"PORT" envDefault:"8080"`
	DomainAllowed string `env:"DOMAIN_ALLOWED"`
}

var (
	cfg Config
)

func appStart() {
	ser := newServer().apiMapping()
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      ser.route,
		ReadTimeout:  3 * time.Minute,
		WriteTimeout: 75 * time.Second,
	}

	log.Print("👉 client work on :", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s", err)
	}
}

func main() {
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	appStart()

}
