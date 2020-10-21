package templates

// EchoStartTmpl : is tmpl to create app
var EchoStartTmpl = `package main

import (
	"database/sql"
	"fmt"
	"log"
	

	"github.com/heraju/mestri/app/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	const (
		host     = "localhost"
		port     = 5432
		user     = "hemanthraju"
		password = "postgress"
		dbname   = "insta_office"
	)

	
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	die(err)
	defer db.Close()

	e := echo.New()
	services.Connect(e, db)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(":" + "1234"))
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
`
