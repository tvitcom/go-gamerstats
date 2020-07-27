package main

import (
	// "encoding/json"
	"context"
	"github.com/cnjack/throttle"
	"github.com/gin-gonic/gin"
	monpagin "github.com/gobeam/mongo-go-pagination"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/acme/autocert"
	"regexp"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	CERTS_CACHE       string
	STORAGE_DRV       string
	STORAGE_DSN       string
	APP_ENTRYPOINT    string
	APP_FQDN          string
	APP_SSLENTRYPOINT string
	API_USER          string
	API_PASSWORD      string
	DB_TYPE           string
	DB_HOST           string
	DB_PORT           string
	DB_NAME           string
	DB_USER           string
	DB_PASSWORD       string
	MONGODB_DSN       string
)

type (
	AnyInterface interface{}
	User         struct {
		Id         primitive.ObjectID `json:"_id" bson:"_id"`
		Email      string             `json:"email" bson:"email"`
		Last_name  string             `json:"last_name" bson:"last_name"`
		Country    string             `json:"country" bson:"country"`
		City       string             `json:"city" bson:"city"`
		Gender     string             `json:"gender" bson:"gender"`
		Birth_date string             `json:"birth_data" bson:"birth_data"`
	}
	UserGames struct {
		ID            primitive.ObjectID `bson:"_id,omitempty"`
		Points_gained string             `bson:"points_gained,omitempty"`
		Win_status    string             `bson:"win_status,omitempty"`
		Game_type     string             `bson:"game_type,omitempty"`
		Created       string             `bson:"created,omitempty"`
	}
	DataOutput struct {
		context    AnyInterface
		pagination AnyInterface
		data       AnyInterface
		errors     AnyInterface
	}
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	gin.SetMode(os.Getenv("GIN_MODE"))
	CERTS_CACHE = os.Getenv("certs_cache") // should be withoud finalize dot
	APP_FQDN = os.Getenv("app_fqdn")       // should be withoud finalize dot
	APP_ENTRYPOINT = os.Getenv("app_entrypoint")
	APP_SSLENTRYPOINT = os.Getenv("app_ssl_entrypoint")
	API_USER = os.Getenv("api_user")
	API_PASSWORD = os.Getenv("api_password")
	DB_TYPE = os.Getenv("db_type")
	DB_HOST = os.Getenv("db_host")
	DB_PORT = os.Getenv("db_port")
	DB_NAME = os.Getenv("db_name")
	DB_USER = os.Getenv("db_user")
	DB_PASSWORD = os.Getenv("db_password")
	MONGODB_DSN = DB_TYPE + "://" + DB_HOST + ":" + DB_PORT
}
func InitMongoDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(MONGODB_DSN))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}
func main() {
	cli := InitMongoDB()
	router := gin.New()
	if gin.Mode() == gin.ReleaseMode {
		fmt.Println("PRODUCTION MODE: Enabled (logs, console, debug messages)")
		gin.DisableConsoleColor()
		f, _ := os.Create("./logs/server.log")
		gin.DefaultWriter = io.MultiWriter(f)
		router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			//custom format for logging:
			return fmt.Sprintf("%s - [%s] %s \"%s\" %d \"%s\" %s\n",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.ClientIP,
				param.Method,
				param.Path,
				param.StatusCode,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		}))
	} else {
		fmt.Println("PRODUCTION MODE: Disabled: api_user,api_password:", API_USER, API_PASSWORD)
	}
	// Define common middlewares
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("templates/*")
	// Server settings
	s := &http.Server{
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 32 << 20,
	}
	s.Addr = APP_ENTRYPOINT

	// Static assets
	router.StaticFile("/favicon.ico", "./assets/img/favicon-32x32.png")
	router.Static("/public", "./public")

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, `Please use our api by link: /api_v1`)
	})

	api_v1 := router.Group("/api_v1", throttle.Policy(&throttle.Quota{
		Limit:  1,
		Within: time.Second,
	}))

	api_v1.Use(gin.BasicAuth(gin.Accounts{
		API_USER: API_PASSWORD,
	}))

	api_v1.GET("/", func(c *gin.Context) {
		bs, err := ioutil.ReadFile("./API.md")
		if err != nil {
			return
		}
		str := string(bs)
		c.HTML(http.StatusOK, "apidoc.htmlt", gin.H{
			"apiver": 1,
			"doc":    str,
		})
	})

	// User part APIv1
	user := api_v1.Group("/user")
	user.GET("/listing", func(c *gin.Context) {
		var status int
		pagenum := c.DefaultQuery("pagenum", "0")
		//validation http keys
		var validStr = regexp.MustCompile(`^[0-9]{1,5}$`)
		if ok := validStr.MatchString(pagenum); !ok {
			status = http.StatusBadRequest
			c.JSON(status, gin.H{})
			return
		}
		// print pagination data
		var limit int64 = 20
		var page int64
		collections := cli.Database(DB_NAME).Collection("users")
		filtr := bson.M{}
		projection := bson.D{
			{"_id", 1},
			{"last_name", 1},
			{"email", 1},
			{"country", 1},
			{"city", 1},
		}
		page, err := strconv.ParseInt(pagenum, 10, 64)
		if err != nil {
			panic(err)
		}
		paginatedData, err := monpagin.New(collections).Limit(limit).Page(page).Select(projection).Filter(filtr).Sort("country", 1).Find()
		if err != nil {
			panic(err)
		}
		var lists []User
		for _, raw := range paginatedData.Data {
			var user *User
			if marshallErr := bson.Unmarshal(raw, &user); marshallErr == nil {
				lists = append(lists, *user)
			}
		}
		// fmt.Printf("DEBUG:Norm Find Data: %+v\n", lists)
		// fmt.Printf("DEBUG:Normal find pagination info: %+v\n", paginatedData.Pagination)
		c.JSON(http.StatusOK, gin.H{
			"context":    "restful",
			"data":       lists,
			"pagination": paginatedData.Pagination,
			"errors":     "",
		})
	})
	user.GET("/profile/:user_id", func(c *gin.Context) {
		var status int
		var result bson.M
		//Validation user_id
		user_id := c.Param("user_id")
		var validStr = regexp.MustCompile(`^[0-9a-f]{24}$`)
		if ok := validStr.MatchString(user_id); !ok {
			status = http.StatusBadRequest
			c.JSON(status, gin.H{})
			return
		}
		//get user info by user_id
		collection := cli.Database(DB_NAME).Collection("users")
		opts := options.FindOne().SetSort(bson.D{{"country", 1}})
		userId, err := primitive.ObjectIDFromHex(user_id)
		filter := bson.M{"_id": userId}
		err = collection.FindOne(context.TODO(), filter, opts).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				status = http.StatusNoContent
				c.JSON(status, gin.H{})
				return
			}
			log.Fatal(err)
		}
		status = http.StatusOK
		fmt.Printf("DEBUG:found document %v", result)
		c.JSON(status, gin.H{
			"context":    "restful," + user_id,
			"data":       result,
			"pagination": "",
			"errors":     "",
		})
	})
	user.GET("/stats/:user_id", func(c *gin.Context) {
		var status int
		user_id := c.Param("user_id")
		pagenum := c.DefaultQuery("pagenum", "0")
		groupingtype := c.DefaultQuery("groupingtype", "by_day") //and by_game
		_ = groupingtype
		//find user by id:
		var userInfo bson.M
		collection := cli.Database(DB_NAME).Collection("users")
		opts := options.FindOne().SetSort(bson.D{{"country", 1}})
		userId, err := primitive.ObjectIDFromHex(user_id)
		filter := bson.M{"_id": userId}
		err = collection.FindOne(context.TODO(), filter, opts).Decode(&userInfo)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				status = http.StatusNoContent
				c.JSON(status, gin.H{})
				return
			}
			log.Fatal(err)
		}
		status = http.StatusOK
		fmt.Printf("DEBUG:found document %v", userInfo)

		//find users statistic:
		var limit int64 = 20
		var page int64
		collections := cli.Database(DB_NAME).Collection("user_games")
		filtr := bson.M{}
		//UserGames struct {
		// ID          primitive.ObjectID `bson:"_id,omitempty"`
		// Points_gained string           `bson:"points_gained,omitempty"`
		// Win_status    string           `bson:"win_status,omitempty"`
		// Game_type     string           `bson:"game_type,omitempty"`
		// Created       string
		projection := bson.D{
			{"_id", 1},
			{"points_gained", 1},
			{"win_status", 1},
			{"game_type", 1},
		}
		page, err = strconv.ParseInt(pagenum, 10, 64)
		if err != nil {
			panic(err)
		}
		paginatedData, err := monpagin.New(collections).Limit(limit).Page(page).Select(projection).Filter(filtr).Sort("country", 1).Find()
		if err != nil {
			panic(err)
		}
		var lists []UserGames
		for _, raw := range paginatedData.Data {
			var userGames *UserGames
			if marshallErr := bson.Unmarshal(raw, &userGames); marshallErr == nil {
				lists = append(lists, *userGames)
			}
		}
		fmt.Printf("DEBUG:Norm Find Data: %+v\n", lists)
		fmt.Printf("DEBUG:Normal find pagination info: %+v\n", paginatedData.Pagination)
		c.JSON(http.StatusOK, gin.H{
			"context":    "restful," + groupingtype,
			"data":       lists,
			"pagination": paginatedData.Pagination,
			"errors":     "",
		})
	})

	//Predefined for errors requests
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, `404 NotFound`)
	})
	router.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, `405 MethodNotAllowed`)
	})

	// Listen and serve:
	if gin.Mode() == gin.ReleaseMode {
		go func() {
			if err := http.ListenAndServe(APP_ENTRYPOINT, http.HandlerFunc(redirectHTTPS)); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()
		tlsManager := &autocert.Manager{
			Cache:      autocert.DirCache(CERTS_CACHE),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(APP_FQDN, "www"+APP_FQDN),
		}
		s := &http.Server{
			Addr:           ":3363",
			TLSConfig:      tlsManager.TLSConfig(),
			Handler:        router,
			ReadTimeout:    60 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 32 << 20,
		}
		s.ListenAndServeTLS("", "")
	} else {
		s.ListenAndServe()
	}
}

func redirectHTTPS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+APP_FQDN+r.RequestURI, http.StatusMovedPermanently)
}
