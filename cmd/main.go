package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

var db *sql.DB
var botInstance *tgbotapi.BotAPI

type User struct {
	ID    int64
	Name  string
	Phone string
}

func main() {
	// PostgreSQL ma'lumotlar bazasi uchun bog'lanish sozlamalari
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "godb"
	dbPassword := "0208"
	dbName := "tgbot"

	// PostgreSQL ma'lumotlar bazasiga bog'lanish
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Ma'lumotlar bazasiga ulanishni tekshirish
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Telegram botini yaratish
	botToken := "6588290150:AAEb0jDtup7apLatgxvWbCHmh2MgWX81_Xg"
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error creating new bot instance: %v", err)
	}

	// Assign the bot instance to the global variable
	botInstance = bot

	// Interrupt va syscall signalini qabul qilish uchun context yaratish
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Updates channel ochish
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	// Updatesni eshitish
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

	if _, err := db.Exec(`INSERT INTO users (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, msg.From.ID); err != nil {
		log.Printf("Error inserting user ID into database: %v", err)
	}

	user := getUserFromDB(msg.From.ID)
	//if user == nil {
	//	// Handle case where user is not found or database query failed
	//	log.Printf("User not found for ID: %d", msg.From.ID)
	//	return
	//}

	var message string
	if user.Name == "" {
		message = "Assalomu alaykum! Botga xush kelibsiz. Ismingizni kiriting, iltimos."
	} else {
		message = "Assalomu alaykum! Botga xush kelibsiz."
	}

	msgSend := tgbotapi.NewMessage(chatID, message)

	// Insert user ID into database if not already present

	if _, err := botInstance.Send(msgSend); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	user := getUserFromDB(userID)

	log.Printf("Received message: %s", msg.Text)

	if user.Name == "" {
		// Foydalanuvchi ismini so'raymiz
		user.Name = msg.Text
		saveUserToDB(user)

		// Ismni so'ragan xabar
		message := "Assalomu alaykum, " + user.Name + "! Endi telefon raqamingizni yuboring."
		msgSend := tgbotapi.NewMessage(chatID, message)

		contactButton := tgbotapi.NewKeyboardButtonContact("Telefon raqamni yuborish")
		replyMarkup := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contactButton))
		msgSend.ReplyMarkup = &replyMarkup

		if _, err := botInstance.Send(msgSend); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	} else if user.Phone == "" {
		// Foydalanuvchidan telefon raqamini olish
		if msg.Contact != nil {
			user.Phone = msg.Contact.PhoneNumber
			saveUserToDB(user)

			message := fmt.Sprintf("Muvaffaqiyatli ro'yxatdan o'tdingiz. Ismingiz: %s, Telefon: %s", user.Name, user.Phone)
			msgSend := tgbotapi.NewMessage(chatID, message)

			// Next step after getting phone number: additional buttons
			queueButton := tgbotapi.NewKeyboardButton("Navbat olish")
			replyMarkup := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(queueButton))
			msgSend.ReplyMarkup = &replyMarkup

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		} else {
			log.Printf("Foydalanuvchidan telefon raqamni kiritish so'raldi, lekin kontakt yuborilmadi")
		}
	} else {
		// Handle other user messages based on the state (after name and phone)
		switch msg.Text {
		case "Navbat olish":
			// User chose to request queue
			message := "Sizning so'rovingiz qabul qilindi. Tez orada sizga aloqaga chiqamiz."
			msgSend := tgbotapi.NewMessage(chatID, message)

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}

			// Implement queue handling logic here
		default:
			// Handle other messages or commands
			// For example, show options, etc.
			message := "Nimani qilmoqchisiz?"
			msgSend := tgbotapi.NewMessage(chatID, message)

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

func getUserFromDB(userID int64) *User {
	var (
		user     User
		nameStr  sql.NullString
		phoneStr sql.NullString
	)

	row := db.QueryRow("SELECT user_id, name, phone FROM users WHERE user_id = $1", userID)
	err := row.Scan(&user.ID, &nameStr, &phoneStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else if err == sql.ErrNoRows {
			log.Printf("User with ID %d not found in database", userID)
			return nil
		} else {
			log.Printf("Error querying user from database: %v", err)
			return nil
		}
	}

	if nameStr.Valid {
		user.Name = nameStr.String
	}

	if phoneStr.Valid {
		user.Phone = phoneStr.String
	}

	return &user
}

func saveUserToDB(user *User) {
	_, err := db.Exec("UPDATE users SET name = $2, phone = $3 WHERE user_id = $1", user.ID, user.Name, user.Phone)
	if err != nil {
		log.Printf("Error updating user in database: %v", err)
	}
}
