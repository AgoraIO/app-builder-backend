package main

import (
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"

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

	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

const defaultPort = "8080"

func main() {
	utils.SetupConfig()

	port := utils.GetPORT(defaultPort)

	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
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
		Debug: viper.GetBool("DEBUG"),
	}))
	middlewareHandler.Use(middleware.AuthHandler(database))
	// middlewareHandler.Use(hlog.AccessHandler())

	if viper.GetBool("ENABLE_NEWRELIC_MONITORING") {
		nrAgent, err := newrelic.NewApplication(
			newrelic.ConfigAppName(viper.GetString("NEWRELIC_APPNAME")),
			newrelic.ConfigLicense(viper.GetString("NEWRELIC_LICENSE")),
			newrelic.ConfigDebugLogger(os.Stdout),
		)

		if err != nil {
			log.Fatal().Err(err).Msg("Error initializing New Relic Agent")
			return
		}

		router.Use(nrgorilla.Middleware(nrAgent))
	}

	middlewareHandler.UseHandler(router)

	log.Debug().Str("PORT", port)
	log.Fatal().Err(http.ListenAndServe(":"+port, middlewareHandler))
}
