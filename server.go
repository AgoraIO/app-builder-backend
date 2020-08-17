package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/oauth"

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
	// router.Use(func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.Header().Set("Access-Control-Allow-Origin", "*")
	// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 		next.ServeHTTP(w, r)
	// 	})
	// })

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
	oauthHandler := oauth.Router{DB: database}

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "success")

		log.Print(w.Header())
	}))
	router.Handle("/query", srv)
	router.Handle("/oauth/web", http.HandlerFunc(oauthHandler.WebOAuthHandler))
	router.Handle("/oauth/desktop", http.HandlerFunc(oauthHandler.DesktopOAuthHandler))
	router.Handle("/oauth/mobile", http.HandlerFunc(oauthHandler.MobileOAuthHandler))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
