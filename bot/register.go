package bot

import (
	"fmt"
	"log"
	"sync"
	"tgbot/config"
	"tgbot/models"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var tempUserData = struct {
	sync.RWMutex
	data map[int64]*models.User
}{data: make(map[int64]*models.User)}

func Register(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection issue")
		return
	}

	user := storage.GetUserFromDB(int64(userID))
	log.Printf("Registering user: %+v", user)

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance issue")
		return
	}

	// Check if the user is already fully registered
	if user != nil && user.Name != "" && user.Phone != "" {
		message := fmt.Sprintf("Assalomu alaykum, %s! Siz allaqachon ro'yxatdan o'tgansiz.", user.Name)
		msgSend := tgbotapi.NewMessage(chatID, message)

		if _, err := botInstance.Send(msgSend); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		SelectBarber(chatID, botInstance)
		return
	}

	// Handle the registration process
	tempUserData.Lock()
	defer tempUserData.Unlock()

	if user == nil {
		user = &models.User{ID: int64(userID)}
	}

	if tempData, exists := tempUserData.data[chatID]; !exists || tempData.Name == "" {
		// Save the name in temporary storage if it hasn't been set yet
		tempUserData.data[chatID] = &models.User{Name: msg.Text}

		// Prompt for phone number
		message := "Assalomu alaykum, " + msg.Text + "! Endi telefon raqamingizni yuboring."
		msgSend := tgbotapi.NewMessage(chatID, message)

		contactButton := tgbotapi.NewKeyboardButtonContact("Telefon raqamni yuborish")
		replyMarkup := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contactButton))
		msgSend.ReplyMarkup = replyMarkup

		if _, err := botInstance.Send(msgSend); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	} else if tempData.Phone == "" {
		// Check if the user sent their contact information
		if msg.Contact != nil {
			// Save the phone number in temporary storage
			tempData.Phone = msg.Contact.PhoneNumber

			// Save the user data to the database
			user.Name = tempData.Name
			user.Phone = tempData.Phone
			storage.SaveUserToDB(user)

			// Confirm registration
			response := fmt.Sprintf("Muvaffaqiyatli ro'yxatdan o'tdingiz. Ismingiz: %s, Telefon: %s", user.Name, user.Phone)
			msgSend := tgbotapi.NewMessage(chatID, response)

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}

			// Proceed to select barber
			SelectBarber(chatID, botInstance)

			// Clean up temporary data
			delete(tempUserData.data, chatID)
		} else {
			// Prompt for phone number again if it was not sent correctly
			message := "Iltimos, telefon raqamingizni yuboring."
			msgSend := tgbotapi.NewMessage(chatID, message)

			contactButton := tgbotapi.NewKeyboardButtonContact("Telefon raqamni yuborish")
			replyMarkup := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contactButton))
			msgSend.ReplyMarkup = replyMarkup

			if _, err := botInstance.Send(msgSend); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}
