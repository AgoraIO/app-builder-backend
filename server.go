package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/routes"

	"github.com/samyak-jain/agora_backend/models"

	"github.com/samyak-jain/agora_backend/utils"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/samyak-jain/agora_backend/graph"
	"github.com/samyak-jain/agora_backend/graph/generated"
)

const defaultPort = "8080"

func main() {
	utils.SetupConfig()

	port := utils.GetPORT()

	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		log.Print(err)
		return
	}

	router := chi.NewRouter()

	router.Use(middleware.AuthHandler(database))
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	}).Handler)

	config := generated.Config{
		Resolvers: &graph.Resolver{DB: database},
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	requestHandler := routes.Router{DB: database}

	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.Handle("/oauth/web", http.HandlerFunc(requestHandler.WebOAuthHandler))
	router.Handle("/oauth/desktop", http.HandlerFunc(requestHandler.DesktopOAuthHandler))
	router.Handle("/oauth/mobile", http.HandlerFunc(requestHandler.MobileOAuthHandler))
	router.Handle("/pstnHandle", http.HandlerFunc(requestHandler.DTMFHandler))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
