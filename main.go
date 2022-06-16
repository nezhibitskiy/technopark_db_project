package main

import (
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"project/pkg/api"
	"project/pkg/api/repository"
)

const PORT = "5000"

func main() {
	db, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	usecase := api.NewUsecase(repo)
	handler := api.NewHandler(usecase)

	fmt.Println("listening port " + PORT)
	log.Fatal(fasthttp.ListenAndServe(":"+PORT, handler.GetHandleFunc()))
}

func NewDB() (*sqlx.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable",
		os.Getenv("DBHOST"), os.Getenv("DBPORT"), os.Getenv("DBUSER"),
		os.Getenv("DBPASSWORD"), os.Getenv("DBNAME"))

	db, err := sqlx.Open("pgx", psqlInfo)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
