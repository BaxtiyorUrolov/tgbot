package config

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"log"
)

var db *sql.DB
var bot *tgbotapi.BotAPI

// InitializeDatabase sets up a connection to PostgreSQL database
func InitializeDatabase(dbHost string, dbPort int, dbUser, dbPassword, dbName string) (*sql.DB, error) {
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	dbConn, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	// Test the connection
	if err = dbConn.Ping(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return dbConn, nil
}

// InitializeBot creates a new Telegram bot instance
func InitializeBot(botToken string) (*tgbotapi.BotAPI, error) {
	botInstance, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("error creating new bot instance: %v", err)
	}

	return botInstance, nil
}

// GetDB returns the initialized database connection
func GetDB() *sql.DB {
	return db
}

// GetBot returns the initialized Telegram bot instance
func GetBot() *tgbotapi.BotAPI {
	return bot
}

// Setup initializes database and bot instances
func Setup(dbHost string, dbPort int, dbUser, dbPassword, dbName, botToken string) error {
	// Initialize database
	dbConn, err := InitializeDatabase(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	db = dbConn

	// Initialize bot
	botInstance, err := InitializeBot(botToken)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %v", err)
	}
	bot = botInstance

	log.Println("Database and bot initialized successfully")
	return nil
}
