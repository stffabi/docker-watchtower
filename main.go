package main // import "github.com/stffabi/watchtower"

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
	"github.com/stffabi/watchtower/actions"
	"github.com/stffabi/watchtower/container"
	"github.com/urfave/cli"
)

var (
	client       container.Client
	scheduleSpec string
	cleanup      bool
	noRestart    bool
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = "watchtower"
	app.Usage = "Automatically update running Docker containers"
	app.Before = before
	app.Action = start
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "daemon socket to connect to",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.IntFlag{
			Name:   "interval, i",
			Usage:  "poll interval (in seconds)",
			Value:  300,
			EnvVar: "WATCHTOWER_POLL_INTERVAL",
		},
		cli.StringFlag{
			Name:   "schedule, s",
			Usage:  "the cron expression which defines when to update",
			EnvVar: "WATCHTOWER_SCHEDULE",
		},
		cli.BoolFlag{
			Name:   "no-pull",
			Usage:  "do not pull new images",
			EnvVar: "WATCHTOWER_NO_PULL",
		},
		cli.BoolFlag{
			Name:   "no-restart",
			Usage:  "do not restart containers",
			EnvVar: "WATCHTOWER_NO_RESTART",
		},
		cli.BoolFlag{
			Name:   "cleanup",
			Usage:  "remove old images after updating",
			EnvVar: "WATCHTOWER_CLEANUP",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug mode with verbose logging",
		},
		cli.StringFlag{
			Name:   "apiversion",
			Usage:  "the version of the docker api",
			EnvVar: "DOCKER_API_VERSION",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	pollingSet := c.IsSet("interval")
	cronSet := c.IsSet("schedule")

	if pollingSet && cronSet {
		log.Fatal("Only schedule or interval can be defined, not both.")
	} else if cronSet {
		scheduleSpec = c.String("schedule")
	} else {
		scheduleSpec = "@every " + strconv.Itoa(c.Int("interval")) + "s"
	}

	cleanup = c.GlobalBool("cleanup")
	noRestart = c.GlobalBool("no-restart")

	// configure environment vars for client
	err := envConfig(c)
	if err != nil {
		return err
	}

	client = container.NewClient(!c.GlobalBool("no-pull"))
	return nil
}

func start(c *cli.Context) error {
	names := c.Args()

	if err := actions.CheckPrereqs(client, cleanup); err != nil {
		log.Fatal(err)
	}

	tryLockSem := make(chan bool, 1)
	tryLockSem <- true

	cron := cron.New()
	err := cron.AddFunc(
		scheduleSpec,
		func() {
			select {
			case v := <-tryLockSem:
				defer func() { tryLockSem <- v }()
				if err := actions.Update(client, names, cleanup, noRestart); err != nil {
					fmt.Println(err)
				}
			default:
				log.Debug("Skipped another update already running.")
			}

			nextRuns := cron.Entries()
			if len(nextRuns) > 0 {
				log.Debug("Scheduled next run: " + nextRuns[0].Next.String())
			}
		})

	if err != nil {
		return err
	}

	log.Info("First run: " + cron.Entries()[0].Schedule.Next(time.Now()).String())
	cron.Start()

	// Graceful shut-down on SIGINT/SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)

	<-interrupt
	cron.Stop()
	log.Info("Waiting for running update to be finished...")
	<-tryLockSem
	os.Exit(1)
	return nil
}

func setEnvOptStr(env string, opt string) error {
	if opt != "" && opt != os.Getenv(env) {
		err := os.Setenv(env, opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func setEnvOptBool(env string, opt bool) error {
	if opt == true {
		return setEnvOptStr(env, "1")
	}
	return nil
}

// envConfig translates the command-line options into environment variables
// that will initialize the api client
func envConfig(c *cli.Context) error {
	var err error

	err = setEnvOptStr("DOCKER_HOST", c.GlobalString("host"))
	err = setEnvOptBool("DOCKER_TLS_VERIFY", c.GlobalBool("tlsverify"))
	err = setEnvOptStr("DOCKER_API_VERSION", c.GlobalString("apiversion"))

	return err
}
