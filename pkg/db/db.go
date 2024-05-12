package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/torrentxok/parchis/pkg/cfg"
)

func ConnectToDB() (*pgx.Conn, error) {
	dbConnectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.ConfigVar.Database.User,
		cfg.ConfigVar.Database.Password, cfg.ConfigVar.Database.DBName)
	db, err := pgx.Connect(context.Background(), dbConnectionString)
	if err != nil {
		log.Print("[ERROR] Error connection to DB: ", err)
		return nil, err
	}
	return db, nil
}
