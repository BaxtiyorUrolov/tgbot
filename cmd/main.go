package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"tgbot/bot"
	"tgbot/config"
	"tgbot/models"
	"tgbot/storage"
)

func main() {

	dbConfig := models.DB{
		Host:     "localhost",
		Port:     5432,
		User:     "godb",
		Password: "0208",
		Name:     "tgbot",
	}
	botToken := "6588290150:AAEb0jDtup7apLatgxvWbCHmh2MgWX81_Xg"

	if err := config.Setup(dbConfig, botToken); err != nil {
		log.Fatalf("Error setting up configuration: %v", err)
	}

	defer func() {
		if err := config.GetDB().Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Fatal("Failed to get bot instance from config")
	}

	// Interrupt and syscall signal handling context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := botInstance.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Error getting updates: %v", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down bot...")
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			switch update.Message.Command() {
			case "start":
				handleStartCommand(update.Message)
			default:
				handleMessage(update.Message)
			}
		}
	}
}

func handleStartCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	if _, err := db.Exec(`INSERT INTO users (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, msg.From.ID); err != nil {
		log.Printf("Error inserting user ID into database: %v", err)
		return
	}

	user := storage.GetUserFromDB(int64(msg.From.ID))

	var message string
	if user.Name == "" {
		message = "Assalomu alaykum! Botga xush kelibsiz. Ismingizni kiriting, iltimos."
	} else {
		message = fmt.Sprintf("Assalomu alaykum, %s! Botga xush kelibsiz.", user.Name)

	}

	msgSend := tgbotapi.NewMessage(chatID, message)

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	if _, err := botInstance.Send(msgSend); err != nil {
		log.Printf("Error sending message: %v", err)
	}

	if user.Name != "" {
		bot.SelectBarber(chatID)
	}
}

func handleMessage(msg *tgbotapi.Message) {

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	//user := storage.GetUserFromDB(userID)

	log.Printf("Received message: %s", msg.Text)

	bot.Register(msg)

}
