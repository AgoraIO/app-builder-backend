package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/samyak-jain/agora_backend/oauth"

	"github.com/samyak-jain/agora_backend/middleware"
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
		log.Panic(err)
	}

	router := chi.NewRouter()
	router.Use(cors.Default().Handler)
	router.Use(middleware.AuthHandler(database))

	config := generated.Config{
		Resolvers: &graph.Resolver{DB: database},
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	oauthHandler := oauth.Router{DB: database}

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.Handle("/oauth/web", http.HandlerFunc(oauthHandler.WebOAuthHandler))
	router.Handle("/oauth/desktop", http.HandlerFunc(oauthHandler.DesktopOAuthHandler))
	router.Handle("/oauth/mobile", http.HandlerFunc(oauthHandler.MobileOAuthHandler))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
