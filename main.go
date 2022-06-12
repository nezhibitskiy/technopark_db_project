package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"log"
	"os"
	"project/internal"
)

func ConnectDB() (*pgxpool.Pool, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable",
		os.Getenv("DBHOST"), os.Getenv("DBPORT"), os.Getenv("DBUSER"),
		os.Getenv("DBPASSWORD"), os.Getenv("DBNAME"))

	dbPool, err := pgxpool.Connect(context.Background(), psqlInfo)
	if err != nil {
		return nil, err
	}

	return dbPool, nil
}

func main() {
	dbPool, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()
	s := echo.New()

	internal.RegisterService(s, dbPool)
	err = s.Start(":5000")
	if err != nil {
		log.Fatal(err)
	}
}
