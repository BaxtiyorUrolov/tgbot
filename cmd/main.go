	package main

	import (
		"context"
		"fmt"
		"log"
		"os"
		"os/signal"
		"strings"
		"tgbot/bot"
		"tgbot/config"
		"tgbot/models"
		"tgbot/storage"
		"time"

		tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	)

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
		botToken := "6588290150:AAEb0jDtup7apLatgxvWbCHmh2MgWX81_Xg"

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
			callback := update.CallbackQuery
			fmt.Println("bu yer callback")
			// Logic for handling callback queries
			handleCallbackQuery(callback)
		} else {
			log.Printf("Received unsupported update type: %T", update)
		}
	}

	func handleInlineQuery(inline *tgbotapi.InlineQuery) {
		// Logic for handling callback queries
		log.Printf("Received callback query: %s", inline.Query)
		// Example: Extract data from callback query and perform relevant actions
	}

	func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
		// Extracting data from callback query
		barberName := callback.Data

		// Printing barber name to the console
		fmt.Println("Selected barber:", barberName)
		botInstance := config.GetBot()
		if botInstance == nil {
			log.Println("Bot instance is nil")
			return
		}

		if strings.Contains(strings.ToLower(barberName), "select_date_") {
			fmt.Println("SSS")
			fmt.Println(callback.Message)
			bot.SelectDate(callback.Message.Chat.ID, botInstance, barberName)
		}

		// You can add further logic here based on the selected barber
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

		fmt.Println(msg.From.ID)

		fmt.Println(user.Name)

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
