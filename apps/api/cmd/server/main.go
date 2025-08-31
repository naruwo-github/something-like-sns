package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/example/something-like-sns/apps/api/internal/adapter/handler/rpc"
	"github.com/example/something-like-sns/apps/api/internal/adapter/repository/mysql"
	"github.com/example/something-like-sns/apps/api/internal/application"
)

func mustGetenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def != "" {
		return def
	}
	log.Fatalf("missing env %s", key)
	return ""
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "X-Tenant", "X-User", "Connect-Protocol-Version"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
	}))

	// Health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// DB connect
	dbHost := mustGetenv("DB_HOST", "127.0.0.1")
	dbPort := mustGetenv("DB_PORT", "3306")
	dbUser := mustGetenv("DB_USER", "app")
	dbPass := mustGetenv("DB_PASS", "pass")
	dbName := mustGetenv("DB_NAME", "sns")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4,utf8", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	e.GET("/dbping", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "down"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "up"})
	})

	// Dependency Injection Wiring
	allowDev := mustGetenv("ALLOW_DEV_HEADERS", "true") == "true"

	// 1. Create repositories (driven/secondary adapters)
	authRepo := mysql.NewAuthRepository(db)
	timelineRepo := mysql.NewTimelineRepository(db)
	reactionRepo := mysql.NewReactionRepository(db)
	dmRepo := mysql.NewDMRepository(db)
	cursorEncoder := mysql.NewCursorEncoder()

	// 2. Create use cases (application core)
	authUsecase := application.NewAuthUsecase(authRepo)
	timelineUsecase := application.NewTimelineUsecase(timelineRepo, cursorEncoder)
	reactionUsecase := application.NewReactionUsecase(reactionRepo)
	dmUsecase := application.NewDMUsecase(dmRepo, cursorEncoder)

	// 3. Create interceptor (shared adapter logic)
	authInterceptor := rpc.NewAuthInterceptor(authUsecase, allowDev)

	// 4. Create handlers (driving/primary adapters)
	tenantHandler := rpc.NewTenantHandler(authUsecase, allowDev)
	timelineHandler := rpc.NewTimelineHandler(timelineUsecase)
	reactionHandler := rpc.NewReactionHandler(reactionUsecase)
	dmHandler := rpc.NewDMHandler(dmUsecase)

	// 5. Mount RPC handlers with interceptors
	path1, h1 := tenantHandler.MountHandler(authInterceptor)
	e.Any(path1+"*", echo.WrapHandler(h1))

	path2, h2 := timelineHandler.MountHandler(authInterceptor)
	e.Any(path2+"*", echo.WrapHandler(h2))

	path3, h3 := reactionHandler.MountHandler(authInterceptor)
	e.Any(path3+"*", echo.WrapHandler(h3))

	path4, h4 := dmHandler.MountHandler(authInterceptor)
	e.Any(path4+"*", echo.WrapHandler(h4))

	port := mustGetenv("API_PORT", "8080")
	log.Printf("API listening on :%s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}