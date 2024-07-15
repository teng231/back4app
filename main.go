package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/teng231/back4app/db"
	"github.com/teng231/back4app/httpserver"
	"github.com/urfave/cli/v2"
)

type Config struct {
	DbDSN         string `env:"DB_DSN"`
	Port          string `env:"PORT" envDefault:"8080"`
	DomainAllowed string `env:"DOMAIN_ALLOWED"`
	BotToken      string `env:"BOT_TOKEN"`
}

var (
	cfg Config
)

func syncdb(ctx *cli.Context) error {
	log.Print(cfg.DbDSN)
	tidb, err := db.New(cfg.DbDSN)
	if err != nil {
		log.Print("db connect fail ", err)
		return err
	}
	if err := tidb.Migrate(); err != nil {
		return err
	}
	log.Print("Tables created")
	return nil
}

func appStart(ctx *cli.Context) error {
	tidb, err := db.New(cfg.DbDSN)
	if err != nil {
		log.Print("db connect fail ", err)
	}
	// ################### REGISTER BOT HANDLERS ###################

	bot := newBot(cfg.BotToken, tidb).registerHandlers()

	go bot.Start()
	log.Print("👉 bot listenning")
	// ################## REGISTER REST HANDLERS ##################

	ser := newServer().customMiddleware().registerHandlers()
	log.Print("👉 client work on :", cfg.Port)
	err = httpserver.HTTPStart(ser.Engine, cfg.Port, 3*time.Minute, 75*time.Second).ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s", err)
	}
	return err
}

func appRoot() error {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) error {
		return errors.New("please enter your command")
	}
	app.Commands = []*cli.Command{
		{Name: "start", Usage: "start up running app", Action: appStart},
		{Name: "syncdb", Usage: "create table in db with struct", Action: syncdb},
	}

	return app.Run(os.Args)
}

func main() {
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if err := appRoot(); err != nil {
		panic(err)
	}
}
