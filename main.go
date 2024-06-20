package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dzfranklin/plantopo-api/analysis"
	"github.com/dzfranklin/plantopo-api/authn"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/dzfranklin/plantopo-api/routes"
	"github.com/dzfranklin/plantopo-api/settings"
	"github.com/dzfranklin/plantopo-api/tracks"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := godotenv.Load(".env", ".env.local")
	if err != nil {
		slog.Info("dotenv", "error", err)
	}

	appEnv := getEnvOr("APP_ENV", "production")

	host := getEnvOr("HOST", "0.0.0.0")
	port := getEnvOr("PORT", "8000")
	addr := host + ":" + port

	dbURL := mustGetEnv("DATABASE_URL")
	pool, err := db.NewPool(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	var authenticator routes.Authenticator
	workosClientID := mustGetEnv("WORKOS_CLIENT_ID")
	workosAPIKey := mustGetEnv("WORKOS_API_KEY")
	authenticator, err = authn.NewWorkOS(workosClientID, workosAPIKey)
	if err != nil {
		log.Fatal(err)
	}
	if appEnv == "development" {
		authenticator = &authn.DevAuthenticator{WorkOS: authenticator.(*authn.WorkOS)}
	}

	toGeoJSONService := tracks.NewToGeoJSONService(mustGetEnv("TO_GEOJSON_SERVICE"))
	elevationService := analysis.NewElevationService(mustGetEnv("ELEVATION_SERVICE"))
	analyzer := analysis.NewAnalyzer(elevationService)

	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)

	workers := river.NewWorkers()
	tracks.AddImportWorker(workers, pool, toGeoJSONService, analyzer)

	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err := riverClient.Start(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	tracksRepo := tracks.NewRepo(pool, riverClient)
	settingsRepo := settings.NewRepo(pool)

	router := routes.Router(
		authenticator,
		tracksRepo,
		elevationService,
		settingsRepo,
	)

	go func() {
		log.Println("Server running on", addr)
		err = router.Run(addr)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		<-sigintOrTerm
		fmt.Printf("Received SIGINT/SIGTERM; initiating soft stop (try to wait for jobs to finish)\n")

		softStopCtx, softStopCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer softStopCtxCancel()

		go func() {
			select {
			case <-sigintOrTerm:
				fmt.Printf("Received SIGINT/SIGTERM again; initiating hard stop (cancel everything)\n")
				softStopCtxCancel()
			case <-softStopCtx.Done():
				fmt.Printf("Soft stop timeout; initiating hard stop (cancel everything)\n")
			}
		}()

		err := riverClient.Stop(softStopCtx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		if err == nil {
			fmt.Printf("Soft stop succeeded\n")
			return
		}

		hardStopCtx, hardStopCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer hardStopCtxCancel()

		// As long as all jobs respect context cancellation, StopAndCancel will
		// always work. However, in the case of a bug where a job blocks despite
		// being cancelled, it may be necessary to either ignore River's stop
		// result (what's shown here) or have a supervisor kill the process.
		err = riverClient.StopAndCancel(hardStopCtx)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Hard stop timeout; ignoring stop procedure and exiting unsafely\n")
		} else if err != nil {
			panic(err)
		}

		// hard stop succeeded
	}()
	<-riverClient.Stopped()
}

func getEnvOr(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	} else {
		return value
	}
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing env var %s", key)
	}
	return value
}
