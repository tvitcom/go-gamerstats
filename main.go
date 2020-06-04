package main

import (
	// "encoding/json"
	"fmt"
	"github.com/cnjack/throttle"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"my.localhost/funny/bitlabs/approot/storage"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	STORAGE_DRV       string
	STORAGE_DSN       string
	APP_ENTRYPOINT    string
	APP_FQDN          string
	APP_SSLENTRYPOINT string
	API_USER          string
	API_PASSWORD      string
)

type ()

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	gin.SetMode(os.Getenv("GIN_MODE"))
	APP_FQDN = os.Getenv("app_fqdn") // should be withoud finalize dot
	APP_ENTRYPOINT = os.Getenv("app_entrypoint")
	APP_SSLENTRYPOINT = os.Getenv("app_ssl_entrypoint")
	API_USER = os.Getenv("api_user")
	API_PASSWORD = os.Getenv("api_password")
	STORAGE_DRV = os.Getenv("db_type")
	STORAGE_DSN = os.Getenv("db_user") + ":" + os.Getenv("db_pass") + "@/" + os.Getenv("db_name")
}

func main() {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		// Sett log format:
		fmt.Println("PRODUCTION MODE: Enabled (logs, console, debug messages)")
	} else {
		fmt.Println("PRODUCTION MODE: Disabled: api_user,api_password:", API_USER, API_PASSWORD)
	}
	router := gin.New()
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

	router.GET("/", func(c *gin.Context) {
		c.String(200, `Please use our api by link: /api_v1`)
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
		pagenum := c.DefaultQuery("pagenum", "0")
		c.String(200, `/listing with page=`+pagenum+` OK!`)
	})
	user.GET("/profile/:user_id", func(c *gin.Context) {
		user_id := c.Param("user_id")
		c.String(200, `/profile for`+user_id+` OK!`)
	})
	user.GET("/stats/:user_id", func(c *gin.Context) {
		user_id := c.Param("user_id")
		pagenum := c.DefaultQuery("pagenum", "0")
		groupingtype := c.DefaultQuery("groupingtype", "by_day") //and by_game
		c.String(200, `stats by `+groupingtype+` for `+user_id+pagenum+` OK!`)
	})

	//Predefined for errors requests
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, `404 NotFound`)
	})
	router.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, `405 MethodNotAllowed`)
	})

	// Listen and serve:
	s.ListenAndServe()
}
