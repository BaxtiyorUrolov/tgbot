package bot

import (
	"strconv"
	"sync"
	"tgbot/config"
	"tgbot/models"
	"tgbot/state"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var tempBarbers = struct {
	sync.RWMutex
	M map[int64]models.Barber
}{M: make(map[int64]models.Barber)}


func HandleAdminStatistics(chatID int64, botInstance *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Statistika: (bu yerda statistikani chiqaring)")
	botInstance.Send(msg)
}

func HandleAdminAddBarber(chatID int64, botInstance *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Barber qo'shish: Iltimos, barberning ID raqamini yuboring.")
	botInstance.Send(msg)

	state.UserStates.Lock()
	state.UserStates.M[chatID] = "adding_barber_id"
	state.UserStates.Unlock()
}

func HandleAdminDeleteBarber(chatID int64, botInstance *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Barber o'chirish: (bu yerda barber o'chirish jarayonini boshqaring)")
	botInstance.Send(msg)
}

func HandleAddBarberID(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "Noto'g'ri ID raqam, iltimos qayta yuboring:")
		config.GetBot().Send(reply)
		return
	}

	tempBarbers.Lock()
	tempBarbers.M[chatID] = models.Barber{ID: id}
	tempBarbers.Unlock()

	reply := tgbotapi.NewMessage(chatID, "Barberning ismini yuboring:")
	config.GetBot().Send(reply)

	state.UserStates.Lock()
	state.UserStates.M[chatID] = "adding_barber_name"
	state.UserStates.Unlock()
}

func HandleAddBarberName(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	tempBarbers.Lock()
	barber := tempBarbers.M[chatID]
	barber.Name = text
	tempBarbers.M[chatID] = barber
	tempBarbers.Unlock()

	reply := tgbotapi.NewMessage(chatID, "Barberning usernamesini yuboring:")
	config.GetBot().Send(reply)

	state.UserStates.Lock()
	state.UserStates.M[chatID] = "adding_barber_username"
	state.UserStates.Unlock()
}

func HandleAddBarberUserName(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	tempBarbers.Lock()
	barber := tempBarbers.M[chatID]
	barber.UserName = text
	tempBarbers.M[chatID] = barber
	tempBarbers.Unlock()

	reply := tgbotapi.NewMessage(chatID, "Barberning telefon raqamini yuboring:")
	config.GetBot().Send(reply)

	state.UserStates.Lock()
	state.UserStates.M[chatID] = "adding_barber_phone"
	state.UserStates.Unlock()
}

func HandleAddBarberPhone(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	tempBarbers.Lock()
	barber := tempBarbers.M[chatID]
	barber.Phone = text
	tempBarbers.Unlock()

	err := storage.AddBarber(barber)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "Barberni qo'shishda xatolik: " + err.Error())
		config.GetBot().Send(reply)
	} else {
		reply := tgbotapi.NewMessage(chatID, "Barber muvaffaqiyatli qo'shildi.")
		config.GetBot().Send(reply)
	}

	state.UserStates.Lock()
	delete(state.UserStates.M, chatID)
	state.UserStates.Unlock()

	tempBarbers.Lock()
	delete(tempBarbers.M, chatID)
	tempBarbers.Unlock()
}
