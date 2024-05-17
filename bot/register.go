//bot/register.go

package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"tgbot/config"
	"tgbot/storage"
)

func Register(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	user := storage.GetUserFromDB(int64(userID))

	if user.Name == "" {
		// Foydalanuvchi ismini so'raymiz
		user.Name = msg.Text
		storage.SaveUserToDB(user)

		// Ismni so'ragan xabar
		message := "Assalomu alaykum, " + user.Name + "! Endi telefon raqamingizni yuboring."
		msgSend := tgbotapi.NewMessage(chatID, message)

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
	} else if user.Phone == "" {
		// Foydalanuvchidan telefon raqamini olish
		if msg.Contact != nil {
			user.Phone = msg.Contact.PhoneNumber
			storage.SaveUserToDB(user)

			// Send success message and provide further options
			response := fmt.Sprintf("Muvaffaqiyatli ro'yxatdan o'tdingiz. Ismingiz: %s, Telefon: %s", user.Name, user.Phone)
			msgSend := tgbotapi.NewMessage(chatID, response)

			// Send initial options after phone number is received

			botInstance := config.GetBot()
			if botInstance == nil {
				log.Println("Bot instance is nil")
				return
			}

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}

			SelectBarber(chatID, botInstance)
		}
	}
}
