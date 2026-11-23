package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"server/internal/auth"
	authv1 "server/internal/auth/v1"
	"server/internal/db"
	"server/internal/middleware"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	port := os.Getenv("SERVER_PORT")

	mux := http.NewServeMux()

	config := huma.DefaultConfig("Nodero API", "1.0.0")
	config.DocsRenderer = huma.DocsRendererScalar
	api := humago.New(mux, config)

	v1 := huma.NewGroup(api, "/api/v1")

	dsn := os.Getenv("DB_DSN")
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to open db", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.SetConnMaxLifetime(3 * time.Minute)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(10)

	db := db.New(conn)

	backendURL := strings.Trim(os.Getenv("BACKEND_URL"), "/")

	googleCfg := &oauth2.Config{
		Scopes:       []string{"openid", "email", "profile"},
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  fmt.Sprintf("%s/api/v1/auth/google/callback", backendURL),
		Endpoint:     google.Endpoint,
	}

	authSvc := auth.NewService(db, googleCfg)
	authv1.NewHandler(authSvc).Register(v1)

	log.Printf("listening http :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LoggerMiddleware(mux)))
}
