package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/learninNdi/gotoko/app/models"
	"github.com/learninNdi/gotoko/database/seeders"
	"github.com/urfave/cli"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	DB        *gorm.DB
	Router    *mux.Router
	AppConfig *AppConfig
}

type AppConfig struct {
	AppName string
	AppURL  string
	AppEnv  string
	AppPort string
}

type DBConfig struct {
	DBDriver   string
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
}

type PageLink struct {
	Page          int32
	Url           string
	IsCurrentPage bool
}

type PaginationLinks struct {
	CurrentPage  string
	NextPage     string
	PreviousPage string
	TotalRows    int32
	TotalPages   int32
	Links        []PageLink
}

type PaginationParams struct {
	Path        string
	TotalRows   int32
	PerPage     int32
	CurrentPage int32
}

type Result struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

var store *sessions.CookieStore
var sessionShoppingCart = "shopping-cart-session"
var sessionFlash = "flash-session"
var sessionUser = "user-session"

func (server *Server) Initialize(appConfig AppConfig, dbConfig DBConfig) {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	fmt.Println("Welcome to " + appConfig.AppName)

	server.initializeAppConfig(appConfig)
	server.InitializeDB(dbConfig)
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Printf("Listening to port %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Router))
}

func (server *Server) InitializeDB(dbConfig DBConfig) {
	var err error

	if dbConfig.DBDriver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName,
		)

		server.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			dbConfig.DBHost, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName, dbConfig.DBPort,
		)

		server.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		panic("Failed on connecting to the database server")
	}
}

func (server *Server) initializeAppConfig(appConfig AppConfig) {
	server.AppConfig = &appConfig
}

func (server *Server) dbMigrate() {
	for _, model := range models.RegisterModels() {
		err := server.DB.Debug().AutoMigrate(model.Model)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database migration successfully.")
}

func (server *Server) InitCommands(dbConfig DBConfig) {
	cmdApp := cli.NewApp()
	cmdApp.Commands = []cli.Command{
		{
			Name: "db:migrate",
			Action: func(c *cli.Context) error {
				server.dbMigrate()

				return nil
			},
		},
		{
			Name: "db:seed",
			Action: func(c *cli.Context) error {
				err := seeders.DBSeed(server.DB)

				if err != nil {
					log.Fatal(err)
				}

				return nil
			},
		},
	}

	err := cmdApp.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func GetPaginationLinks(appConfig *AppConfig, params PaginationParams) (PaginationLinks, error) {
	var links []PageLink

	// total page
	totalPages := int32(math.Ceil(float64(params.TotalRows) / float64(params.PerPage)))

	for i := 1; int32(i) <= totalPages; i++ {
		links = append(links, PageLink{
			Page:          int32(i),
			Url:           fmt.Sprintf("%s/%s?page=%s", appConfig.AppURL, params.Path, fmt.Sprint(i)),
			IsCurrentPage: int32(i) == params.CurrentPage,
		})
	}

	var nextPage int32
	var prevPage int32

	prevPage = 1
	nextPage = totalPages

	if params.CurrentPage > 2 {
		prevPage = params.CurrentPage - 1
	}

	if params.CurrentPage < totalPages {
		nextPage = params.CurrentPage + 1
	}

	return PaginationLinks{
		CurrentPage:  fmt.Sprintf("%s/%s?page=%s", appConfig.AppURL, params.Path, fmt.Sprint(params.CurrentPage)),
		NextPage:     fmt.Sprintf("%s/%s?page=%s", appConfig.AppURL, params.Path, fmt.Sprint(nextPage)),
		PreviousPage: fmt.Sprintf("%s/%s?page=%s", appConfig.AppURL, params.Path, fmt.Sprint(prevPage)),
		TotalRows:    params.TotalRows,
		TotalPages:   totalPages,
		Links:        links,
	}, nil
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := store.Get(r, sessionFlash)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session.AddFlash(value, name)
	session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) []string {
	session, err := store.Get(r, sessionFlash)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fm := session.Flashes(name)

	if len(fm) == 0 {
		return nil
	}

	session.Save(r, w)

	var flashes []string

	for _, fl := range fm {
		flashes = append(flashes, fl.(string))
	}

	return flashes
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, sessionUser)

	return session.Values["id"] != nil
}

// func ComparePassword(password string, hashedPassword string) bool {
func ComparePassword(password string, savedPassword string) bool {
	// return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil

	return password == savedPassword
}

func (server *Server) CurrentUser(w http.ResponseWriter, r *http.Request) *models.User {
	if !IsLoggedIn(r) {
		return nil
	}

	session, _ := store.Get(r, sessionUser)

	userModel := models.User{}
	user, err := userModel.FindByID(server.DB, session.Values["id"].(string))

	if err != nil {
		session.Values["id"] = nil
		session.Save(r, w)

		return nil
	}

	return user
}
