package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewLogger() *log.Logger {
	logger := log.New(os.Stdout, "", 0)
	logger.Println("Executing new Logger")
	return logger
}

func NewHandler (logger *log.Logger) (http.Handler, error){

	logger.Println("Executing new Handler")

	return http.HandlerFunc(func (http.ResponseWriter, *http.Request){
		logger.Println("Got a request")
	}), nil
}

func NewMux (lc fx.Lifecycle, logger *log.Logger) *http.ServeMux{

	logger.Println("Executing new Mux")

	mux := http.NewServeMux()
	server := &http.Server {
		Addr: ":8080",
		Handler: mux,
	}
	lc.Append(fx.Hook{
		OnStart: func (context.Context) error {
			logger.Println("Starting HTTP server")
			go server.ListenAndServe()
			return nil
		},
		OnStop: func (ctx context.Context) error {
			logger.Println("Stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
	return mux
}

func Register(mux *http.ServeMux, h http.Handler) {
	mux.Handle("/", h)
}

func main() {
	app := fx.New(
		fx.Provide(
			NewLogger,
			NewHandler,
			NewMux,
		),
		fx.Invoke(Register),
		fx.WithLogger(
			func () fxevent.Logger {
				return fxevent.NopLogger
			},
		),
	)
	startCtx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}
	if _, err := http.Get("http://localhost:8080/"); err != nil {
		log.Fatal(err)
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}