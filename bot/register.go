package bot

import (
	"log"
	"sync"
	"tgbot/config"
	"tgbot/models"
	"tgbot/state"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var tempUserData = struct {
	sync.RWMutex
	data map[int64]*models.User
}{data: make(map[int64]*models.User)}

func Register(chatID int64, botInstance *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Ismingizni kiriting, iltimos.")
	_, err := botInstance.Send(msg)
	if err != nil {
		log.Printf("Ism so'rovini yuborishda xatolik: %v", err)
		return
	}
	log.Printf("Foydalanuvchidan ism so'ralmoqda: %d", chatID)

	state.UserStates.Lock()
	state.UserStates.M[chatID] = "register_name"
	log.Printf("Foydalanuvchi holati yangilandi: %d -> %s", chatID, "register_name")
	state.UserStates.Unlock()
}

func HandleRegister(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	state.UserStates.RLock()
	currentState := state.UserStates.M[chatID]
	state.UserStates.RUnlock()

	log.Printf("HandleRegister called for chat ID %d with state: %s and text: %s", chatID, currentState, text)

	switch currentState {
	case "register_name":
		log.Printf("Registering name for chat ID %d: %s", chatID, text)
		tempUserData.Lock()
		tempUserData.data[chatID] = &models.User{ID: chatID, Name: text}
		tempUserData.Unlock()

		reply := tgbotapi.NewMessage(chatID, "Telefon raqamingizni kiriting, iltimos.")
		reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonContact("Telefon raqamni ulashish"),
			),
		)

		_, err := config.GetBot().Send(reply)
		if err != nil {
			log.Printf("Telefon raqamini so'rashda xatolik: %v", err)
			return
		}

		state.UserStates.Lock()
		state.UserStates.M[chatID] = "register_phone"
		log.Printf("Foydalanuvchi holati yangilandi: %d -> %s", chatID, "register_phone")
		state.UserStates.Unlock()

	case "register_phone":
		if msg.Contact != nil {
			phoneNumber := msg.Contact.PhoneNumber
			log.Printf("Registering phone for chat ID %d: %s", chatID, phoneNumber)
			tempUserData.Lock()
			tempUserData.data[chatID].Phone = phoneNumber
			user := tempUserData.data[chatID]
			tempUserData.Unlock()

			storage.SaveUserToDB(user)
			

			// Clear user state and temporary data
			state.UserStates.Lock()
			delete(state.UserStates.M, chatID)
			log.Printf("Foydalanuvchi holati o'chirildi: %d", chatID)
			state.UserStates.Unlock()

			tempUserData.Lock()
			delete(tempUserData.data, chatID)
			tempUserData.Unlock()

			// Send confirmation message
			msg := tgbotapi.NewMessage(chatID, "Ro'yxatdan muvaffaqiyatli o'tdingiz.")
			_, err := config.GetBot().Send(msg)
			if err != nil {
				log.Printf("Tasdiq xabarini yuborishda xatolik: %v", err)
				return
			}

			// Select barber after registration
			SelectBarber(chatID, config.GetBot())
		} else {
			msg := tgbotapi.NewMessage(chatID, "Telefon raqamingizni ulashish uchun tugmani bosing.")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButtonContact("Telefon raqamni ulashish"),
				),
			)
			config.GetBot().Send(msg)
		}
	}
}
