package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func mainInternal() error {
	bs, err := ioutil.ReadFile("your-key.pem")
	if err != nil {
		return nil
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(bs)
	if err != nil {
		return nil
	}
	var appID int64 = 10
	r := http.NewServeMux()
	r.Handle("/", NewHandler(appID, key))
	server := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
		Addr:         ":8080",
	}
	log.Printf("Server is started. Go to http://localhost:%d", 8080)
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := mainInternal(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
