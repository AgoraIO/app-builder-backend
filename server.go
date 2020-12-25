package main

import (
	"fmt"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/slack"
	"github.com/samyak-jain/agora_backend/migrations"
	"net/http"
	"net/http/httputil"
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

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const defaultPort = "8050"

func main() {
	utils.SetupConfig()

	port := utils.GetPORT(defaultPort)

	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing database")
		return
	}

	if viper.GetBool("RUN_MIGRATION") {
		migrations.RunMigration()
	}

	router := mux.NewRouter()

	config := generated.Config{
		Resolvers: &graph.Resolver{DB: database},
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	requestHandler := routes.Router{DB: database}

	//OAuth Setup
	goth.UseProviders(
		google.New(viper.GetString("GOOGLE_CLIENT_ID"), viper.GetString("GOOGLE_CLIENT_SECRET"), viper.GetString("BACKEND_URL")+"/oauth/google/web", "email", "profile"),
		microsoftonline.New(viper.GetString("MICROSOFT_KEY"), viper.GetString("MICROSOFT_SECRET"), viper.GetString("BACKEND_URL")+"/oauth/microsoftonline/web", "email", "profile"),
		apple.New(viper.GetString("APPLE_KEY"), viper.GetString("APPLE_SECRET"), viper.GetString("BACKEND_URL")+"/oauth/apple/web", nil, apple.ScopeName, apple.ScopeEmail),
		slack.New(viper.GetString("SLACK_KEY"), viper.GetString("SLACK_SECRET"), viper.GetString("BACKEND_URL")+"/oauth/slack/callback", "email", "profile"),
	)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", playground.Handler("GraphQL playground", "/query"))
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Error().Err(err).Msg("Error reading request")
			return
		}
		log.Info().Interface("request", string(requestDump)).Msg("Request Details")
	})
	router.Handle("/query", srv)
	router.HandleFunc("/oauth/{provider}/web", requestHandler.WebOAuthHandler)
	router.HandleFunc("/oauth/{provider}/desktop", requestHandler.DesktopOAuthHandler)
	router.HandleFunc("/oauth/{provider}/mobile", requestHandler.MobileOAuthHandler)
	router.HandleFunc("/pstnConfig", requestHandler.PSTNConfig)
	router.HandleFunc("/pstnHandle", requestHandler.DTMFHandler)

	middlewareHandler := negroni.Classic()
	middlewareHandler.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{utils.GetAllowedOrigin()},
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		Debug:            false,
	}))
	middlewareHandler.Use(middleware.AuthHandler(database))

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
	fmt.Println("Listening to PORT :", port)
	log.Debug().Str("PORT", port)
	log.Fatal().Err(http.ListenAndServe(":"+port, middlewareHandler))
}
