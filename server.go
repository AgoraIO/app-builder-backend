package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/rs/cors"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/routes"

	"github.com/samyak-jain/agora_backend/models"

	"github.com/samyak-jain/agora_backend/utils"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
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

	router := mux.NewRouter()

	config := generated.Config{
		Resolvers: &graph.Resolver{DB: database},
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	requestHandler := routes.Router{DB: database}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/oauth/web", requestHandler.WebOAuthHandler)
	router.HandleFunc("/oauth/desktop", http.HandlerFunc(requestHandler.DesktopOAuthHandler))
	router.HandleFunc("/oauth/mobile", http.HandlerFunc(requestHandler.MobileOAuthHandler))
	router.HandleFunc("/pstnHandle", http.HandlerFunc(requestHandler.DTMFHandler))

	middlewareHandler := negroni.Classic()
	middlewareHandler.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	}))
	middlewareHandler.Use(middleware.AuthHandler(database))
	middlewareHandler.UseHandler(router)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, middlewareHandler))
}
