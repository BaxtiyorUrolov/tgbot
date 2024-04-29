package models

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

var BotToken string
