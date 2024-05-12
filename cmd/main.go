package main

import (
	_ "database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/torrentxok/parchis/pkg/auth"
	"github.com/torrentxok/parchis/pkg/cfg"
)

func main() {
	configFile, err := os.Open("cfg/appconfig.json")
	if err != nil {
		log.Fatal("[ERROR] Error opening config file: ", err)
	}
	defer configFile.Close()
	log.Print("[INFO] Config opened!")

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&cfg.ConfigVar)
	if err != nil {
		log.Fatal("[ERROR] Error decoding config file: ", err)
	}
	log.Print("[INFO] Decode config done!")

	logFile, err := os.OpenFile(cfg.ConfigVar.Logging.OutFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print("[INFO] Start logging.")

	log.Print("[INFO] Server start!")
	r := mux.NewRouter()

	r.HandleFunc("/signup", auth.SignUpHandler)
	r.HandleFunc("/confirm_email", auth.ConfirmEmailHandler)
	r.HandleFunc("/login", auth.Login)

	log.Fatal(http.ListenAndServe(":8080", r))
}
