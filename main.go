package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"vayer-electric-backend/db"
	"vayer-electric-backend/env"
	"vayer-electric-backend/gracefulserver"
	"vayer-electric-backend/handler"
	"vayer-electric-backend/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

var (
	httpPort = env.PORT
	log      = logging.GetLogger()
)

// Returns a context to be the main context of the app.
// It's canceled once a syscall.SIGINT or syscall.SIGTERM is received.
// We could use signal.NotifyContext instead but we won't be able to use the actual signal received for loging & debugging
func getMainContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		exitSignal := <-signals
		log.Info(fmt.Sprintf("received %s shutting down", exitSignal.String()))

		cancel()
	}()

	return ctx
}

func main() {
	// TODO: add a adecuate logger

	defer func() { log.Info("bye") }()

	mainCtx := getMainContext()

	src := db.GetDbSource()
	err := src.Migrate("./migrations")

	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	server := gracefulserver.New(&http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: r,
	})

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Get("/", handler.GetProducts())
			r.Get("/{id}", handler.GetProductById())
			r.Post("/", handler.CreateProduct())
			r.Put("/{id}", handler.UpdateProduct())
			r.Delete("/{id}", handler.DeleteProduct())
			r.Get("/category/{id}", handler.GetProductsByCategoryId())
			r.Get("/category/{name}", handler.GetProductsByCategoryName())
			r.Get("/{name}", handler.GetProductByName())
		})
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", handler.GetCategories())
			r.Get("/{id}", handler.GetCategoryById())
			r.Post("/", handler.CreateCategory())
			r.Put("/{id}", handler.UpdateCategory())
			r.Delete("/{id}", handler.DeleteCategory())
		})
		r.Route("/subcategories", func(r chi.Router) {
			r.Get("/", handler.GetSubcategories())
			r.Get("/{id}", handler.GetSubcategoryById())
			r.Post("/", handler.CreateSubcategory())
			r.Put("/{id}", handler.UpdateSubcategory())
			r.Delete("/{id}", handler.DeleteSubcategory())
		})
		r.Route("/images", func(r chi.Router) {
			r.Get("/{name}", handler.ServeProductImage())
		})

	})

	if err := server.StartListening(mainCtx); err != nil {
		log.Error("error starting server", zap.Error(err))
		return
	}

	log.Info(fmt.Sprintf("server listening on port %d", httpPort))

	defer func() {
		log.Info("stopping server")
		server.Shutdown()
		log.Info("server stopped")
	}()

	<-mainCtx.Done()

}
