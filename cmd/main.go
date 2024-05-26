package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"tgbot/bot"
	"tgbot/config"
	"tgbot/models"
	"tgbot/storage"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var userStates = struct {
	sync.RWMutex
	m map[int64]string
}{m: make(map[int64]string)}

func main() {
	// Ma'lumotlar omboriga ulanish uchun konfiguratsiyani sozlash
	dbConfig := models.DB{
		Host:     "localhost",
		Port:     5432,
		User:     "godb",
		Password: "0208",
		Name:     "tgbot",
	}
	// Bot tokenini olish
	botToken := "6902655696:AAGLciESTPSVwmWZlxz8fkrY0EC-Fl7qo_I"

	// Konfiguratsiyani sozlash va boshlash
	if err := config.Setup(dbConfig, botToken); err != nil {
		log.Fatalf("Konfiguratsiyani sozlashda xatolik: %v", err)
	}

	// Ma'lumotlar omborini yopish
	defer func() {
		if err := config.GetDB().Close(); err != nil {
			log.Printf("Ma'lumotlar omborini yopishda xatolik: %v", err)
		}
	}()

	// Bot instansiyasini olish
	botInstance := config.GetBot()
	if botInstance == nil {
		log.Fatal("Bot instansiyasini olishda xatolik")
	}

	// Interrupt va syscall signal qabul qilish uchun kontekstni sozlash
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Bot yangiliklarini qabul qilish uchun botning GetUpdates funksiyasidan foydalanish
	offset := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Botni yopish...")
			return
		default:
			updates, err := botInstance.GetUpdates(tgbotapi.NewUpdate(offset))
			if err != nil {
				log.Printf("Yangiliklarni olishda xatolik: %v", err)
				time.Sleep(5 * time.Second) // Agar xatolik bo'lsa, birlamchi vaqt tanlang
				continue
			}
			// Yangiliklarni boshqarish
			for _, update := range updates {
				HandleUpdate(update)
				offset = update.UpdateID + 1
			}
		}
	}
}

func HandleUpdate(update tgbotapi.Update) {
	// Handle the update based on its type
	if update.Message != nil {
		fmt.Println("bu yer start")
		// Logic for handling regular messages
		handleStartCommand(update.Message)
	} else if update.InlineQuery != nil {
		// Logic for handling inline queries
		fmt.Println("bu yer inline")
		handleInlineQuery(update.InlineQuery)
	} else if update.CallbackQuery != nil {
		fmt.Println("bu yer callback")
		// Logic for handling callback queries
		handleCallbackQuery(update)
	} else {
		log.Printf("Received unsupported update type: %T", update)
	}
}

func handleInlineQuery(inline *tgbotapi.InlineQuery) {
	// Logic for handling inline queries
	log.Printf("Received inline query: %s", inline.Query)
	// Example: Extract data from inline query and perform relevant actions
}

func handleCallbackQuery(update tgbotapi.Update) {
	callback := update.CallbackQuery
	data := callback.Data
	botInstance := config.GetBot()

	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	chatID := callback.Message.Chat.ID

	// Check if the callback data contains "select_date_" to identify barber selection
	if strings.HasPrefix(data, "select_date_") {
		barberName := strings.TrimPrefix(data, "select_date_")

		// Store the barber name in the user state map
		userStates.Lock()
		userStates.m[chatID] = barberName
		userStates.Unlock()

		bot.SelectDate(callback.Message.Chat.ID, botInstance, barberName, update)
	} else if strings.HasPrefix(data, "datte_") {
		// Check if the callback data contains "datte_" to identify date selection

		// Retrieve the barber name from the user state map
		userStates.RLock()
		barberName := userStates.m[chatID]
		userStates.RUnlock()

		bot.SelectOrder(callback.Message.Chat.ID, botInstance, barberName, update)
	} else {
		log.Printf("Unknown callback data: %s", data)
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
		if msg.Text != "/start" {
			handleMessage(msg)
		}
		
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
		bot.SelectBarber(chatID, botInstance)
	}
}

func handleMessage(msg *tgbotapi.Message) {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	log.Printf("Received message: %s", msg.Text)
	bot.Register(msg)
}
