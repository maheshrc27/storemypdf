package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/joho/godotenv"
	"github.com/maheshrc27/storemypdf/internal/database"
	"github.com/maheshrc27/storemypdf/internal/env"
	"github.com/maheshrc27/storemypdf/internal/smtp"
	"github.com/maheshrc27/storemypdf/internal/version"
	"github.com/robfig/cron"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	err = run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL  string
	httpPort int
	cookie   struct {
		secretKey string
	}
	db struct {
		dsn string
	}
	notifications struct {
		email string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type application struct {
	config config
	db     *database.DB
	logger *slog.Logger
	mailer *smtp.Mailer
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:4444")
	cfg.httpPort = env.GetInt("HTTP_PORT", 4444)
	cfg.cookie.secretKey = env.GetString("COOKIE_SECRET_KEY", os.Getenv("SECRET_KEY"))
	cfg.db.dsn = env.GetString("DB_DSN", "mydb.db")
	cfg.notifications.email = env.GetString("NOTIFICATIONS_EMAIL", "")
	cfg.smtp.host = env.GetString("SMTP_HOST", os.Getenv("SMTP_HOST"))
	cfg.smtp.port = env.GetInt("SMTP_PORT", 587)
	cfg.smtp.username = env.GetString("SMTP_USERNAME", os.Getenv("SMTP_USERNAME"))
	cfg.smtp.password = env.GetString("SMTP_PASSWORD", os.Getenv("SMTP_PASSWORD"))
	cfg.smtp.from = env.GetString("SMTP_FROM", "storemypdf <no_reply@storemypdf.com>")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
		mailer: mailer,
	}

	c := cron.New()
	c.AddFunc("@every 00h00m30s", app.DeleteFileScheduler)
	c.Start()

	return app.serveHTTP()
}
