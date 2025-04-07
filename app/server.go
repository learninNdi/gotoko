package app

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/learninNdi/gotoko/app/controllers"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func Run() {
	var server = controllers.Server{}
	var appConfig = controllers.AppConfig{}
	var dbConfig = controllers.DBConfig{}

	// read env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error on loading env file")
	}

	appConfig.AppName = getEnv("APP_NAME", "GoToko")
	appConfig.AppEnv = getEnv("APP_ENV", "development")
	appConfig.AppPort = getEnv("APP_PORT", "9000")

	dbConfig.DBDriver = getEnv("DB_DRIVER", "postgres")
	dbConfig.DBHost = getEnv("DB_HOST", "localhost")
	dbConfig.DBUser = getEnv("DB_USER", "postgres")
	dbConfig.DBPassword = getEnv("DB_PASSWORD", "root")
	dbConfig.DBName = getEnv("DB_NAME", "gotoko_db")
	dbConfig.DBPort = getEnv("DB_PORT", "5432")

	// receive flag
	flag.Parse()
	arg := flag.Arg(0)

	server.InitializeDB(dbConfig)

	if arg != "" {
		server.InitCommands(dbConfig)
	} else {
		server.Initialize(appConfig, dbConfig)
		server.Run(":" + appConfig.AppPort)
	}
}
