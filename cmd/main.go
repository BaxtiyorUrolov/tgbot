package main

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"tgbot/bot"
	"tgbot/config"
)

var db *sql.DB
var botInstance *tgbotapi.BotAPI

type User struct {
	ID    int64
	Name  string
	Phone string
}

func main() {
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "godb"
	dbPassword := "0208"
	dbName := "tgbot"
	botToken := "6588290150:AAEb0jDtup7apLatgxvWbCHmh2MgWX81_Xg"

	// Setup database and bot
	if err := config.Setup(dbHost, dbPort, dbUser, dbPassword, dbName, botToken); err != nil {
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

	updates := botInstance.GetUpdatesChan(u)

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

	fmt.Println("Bu yer")

	if _, err := db.Exec(`INSERT INTO users (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, msg.From.ID); err != nil {
		log.Printf("Error inserting user ID into database: %v", err)
		return
	}

	fmt.Println("Bu yer past")

	user := getUserFromDB(msg.From.ID)

	fmt.Println(user.ID)

	var message string
	if user.Name == "" {
		message = "Assalomu alaykum! Botga xush kelibsiz. Ismingizni kiriting, iltimos."
	} else {
		message = fmt.Sprintf("Assalomu alaykum, %s! Botga xush kelibsiz.", user.Name)
		bot.SendInitialOptions(chatID)
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
}

func handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	user := getUserFromDB(userID)

	log.Printf("Received message: %s", msg.Text)

	if user.Name == "" {
		// Foydalanuvchi ismini so'raymiz
		user.Name = msg.Text
		saveUserToDB(user)

		// Ismni so'ragan xabar
		message := "Assalomu alaykum, " + user.Name + "! Endi telefon raqamingizni yuboring."
		msgSend := tgbotapi.NewMessage(chatID, message)

		fmt.Println("ism qabul qilindi")

		contactButton := tgbotapi.NewKeyboardButtonContact("Telefon raqamni yuborish")
		replyMarkup := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contactButton))
		msgSend.ReplyMarkup = &replyMarkup

		botInstance := config.GetBot()
		if botInstance == nil {
			log.Println("Bot instance is nil")
			return
		}

		if _, err := botInstance.Send(msgSend); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		fmt.Println("tel!!!!!!!!!!!!!!")
	} else if user.Phone == "" {
		// Foydalanuvchidan telefon raqamini olish
		if msg.Contact != nil {
			user.Phone = msg.Contact.PhoneNumber
			saveUserToDB(user)

			// Send success message and provide further options
			response := fmt.Sprintf("Muvaffaqiyatli ro'yxatdan o'tdingiz. Ismingiz: %s, Telefon: %s", user.Name, user.Phone)
			msgSend := tgbotapi.NewMessage(chatID, response)

			bot.SendInitialOptions(chatID) // Send initial options after phone number is received

			botInstance := config.GetBot()
			if botInstance == nil {
				log.Println("Bot instance is nil")
				return
			}

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		} else {
			log.Printf("Foydalanuvchidan telefon raqamni kiritish so'raldi, lekin kontakt yuborilmadi")
		}
	} else {
		// Handle other user messages based on the state (after name and phone)
		switch msg.Text {
		case "Aloqa":
			// User chose to request assistance
			assistanceMessage := "Assalomu alaykum, yordamim kerak bo'lsa +998931792908 raqamiga qo'ng'iroq qiling."
			assistanceMsgSend := tgbotapi.NewMessage(chatID, assistanceMessage)

			botInstance := config.GetBot()

			if _, err := botInstance.Send(assistanceMsgSend); err != nil {
				log.Printf("Error sending assistance message: %v", err)
			}

		case "Navbat olish":
			// User chose to request queue
			queueMessage := "Sizning so'rovingiz qabul qilindi. Tez orada sizga aloqaga chiqamiz."
			queueMsgSend := tgbotapi.NewMessage(chatID, queueMessage)

			botInstance := config.GetBot()

			if _, err := botInstance.Send(queueMsgSend); err != nil {
				log.Printf("Error sending queue message: %v", err)
			}

			// Implement queue handling logic here

		default:
			// Handle other messages or commands
			// For example, show options, etc.
			message := "Nimani qilmoqchisiz?"
			msgSend := tgbotapi.NewMessage(chatID, message)

			botInstance := config.GetBot()

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

	fmt.Println("User: ", userID) // For debugging: print the userID being fetched

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return nil
	}

	row := db.QueryRow("SELECT user_id, name, phone FROM users WHERE user_id = $1", userID)
	err := row.Scan(&user.ID, &nameStr, &phoneStr)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying user from database: %v", err)
		} else {
			log.Printf("User with ID %d not found in database", userID)
		}
		return nil // Return nil if user not found or other error occurred
	}

	fmt.Println("id:", user.ID) // For debugging: print the user ID after scan

	if nameStr.Valid {
		user.Name = nameStr.String
	}

	if phoneStr.Valid {
		user.Phone = phoneStr.String
	}

	return &user // Return pointer to populated User struct
}

func saveUserToDB(user *User) {

	fmt.Println("Saving user: ", user.Name)

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	_, err := db.Exec("UPDATE users SET name = $2, phone = $3 WHERE user_id = $1", user.ID, user.Name, user.Phone)
	if err != nil {
		log.Printf("Error updating user in database: %v", err)
	}
}
