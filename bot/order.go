package bot

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"tgbot/config"
	"tgbot/models"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var userStates = struct {
	sync.RWMutex
	m map[int64]int
}{m: make(map[int64]int)}

func SelectBarber(chatID int64, botInstance *tgbotapi.BotAPI) {
	photoFilePath := "./photo.jpg"
	caption := "Iltimos, sartaroshni tanlang:"

	photoFileBytes, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Rasmni o'qishda xatolik: %v", err)
		return
	}

	barbers, err := storage.GetBarbers()
	if err != nil {
		log.Printf("Sartaroshlarni olishda xatolik: %v", err)
		return
	}

	if len(barbers) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Sartaroshlar mavjud emas.")
		botInstance.Send(msg)
		return
	}

	msg := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoFileBytes,
	})
	msg.Caption = caption

	var rows [][]tgbotapi.InlineKeyboardButton
	switch len(barbers) {
	case 1:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barbers[0].Name, fmt.Sprintf("select_date_%s", barbers[0].Name)),
		))
	case 2:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barbers[0].Name, fmt.Sprintf("select_date_%s", barbers[0].Name)),
			tgbotapi.NewInlineKeyboardButtonData(barbers[1].Name, fmt.Sprintf("select_date_%s", barbers[1].Name)),
		))
	case 3:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barbers[0].Name, fmt.Sprintf("select_date_%s", barbers[0].Name)),
			tgbotapi.NewInlineKeyboardButtonData(barbers[1].Name, fmt.Sprintf("select_date_%s", barbers[1].Name)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barbers[2].Name, fmt.Sprintf("select_date_%s", barbers[2].Name)),
		))
	default:
		for i := 0; i < len(barbers); i += 2 {
			if i+1 < len(barbers) {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(barbers[i].Name, fmt.Sprintf("select_date_%s", barbers[i].Name)),
					tgbotapi.NewInlineKeyboardButtonData(barbers[i+1].Name, fmt.Sprintf("select_date_%s", barbers[i+1].Name)),
				))
			} else {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(barbers[i].Name, fmt.Sprintf("select_date_%s", barbers[i].Name)),
				))
			}
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = &keyboard

	sentMessage, err := botInstance.Send(msg)
	if err != nil {
		log.Printf("Rasmni yuborishda xatolik: %v", err)
	}

	log.Println("Sartarosh tanlash tugmalari yuborildi")

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}


func SelectDate(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, prevMessageID int) {
	if barberName == "" {
		log.Println("Sartarosh tanlanmadi. Sana tanlashni o'tkazib yuborish.")
		return
	}

	deleteMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteMessage)

	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)
	dayAfterTomorrow := today.AddDate(0, 0, 2)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(today.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, today.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(tomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, tomorrow.Format("2006-01-02"))),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dayAfterTomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, dayAfterTomorrow.Format("2006-01-02"))),
		),
	)

	dateSelectionMsg := tgbotapi.NewMessage(chatID, "Iltimos, sana tanlang (kun.oy.yil):")
	dateSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(dateSelectionMsg)
	if err != nil {
		log.Printf("Sana tanlash klaviaturasini yuborishda xatolik: %v", err)
		return
	}

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func SelectOrder(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, orderDate string, update tgbotapi.Update, prevMessageID int) {
	deleteMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteMessage)

	order := models.GetOrders{
		BarberID: barberName,
		Date:     orderDate,
	}

	existingOrderTimes, err := storage.GetOrders(order)
	if err != nil {
		log.Printf("Buyurtmalarni olishda xatolik: %v", err)
		return
	}

	existingTimesSet := make(map[string]struct{})
	for _, time := range existingOrderTimes {
		existingTimesSet[time] = struct{}{}
	}

	timeSlots := []string{"9:00", "10:00", "11:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00"}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(timeSlots); i += 4 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 4 && i+j < len(timeSlots); j++ {
			timeSlot := timeSlots[i+j]
			if _, exists := existingTimesSet[timeSlot]; exists {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("❌", "X"))
			} else {
				callbackData := fmt.Sprintf("confirm_%s_%s_%s", barberName, orderDate, timeSlot)
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(timeSlot, callbackData))
			}
		}
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	timeSelectionMsg := tgbotapi.NewMessage(chatID, "Iltimos, vaqt tanlang:")
	timeSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(timeSelectionMsg)
	if err != nil {
		log.Printf("Vaqt tanlash klaviaturasini yuborishda xatolik: %v", err)
		return
	}

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func HandleConfirmation(chatID int64, botInstance *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, update tgbotapi.Update) {
	callbackData := callback.Data

	deleteMessage := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	botInstance.Send(deleteMessage)

	if strings.HasPrefix(callbackData, "confirm_") {
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Tasdiqlash uchun noto'g'ri callback ma'lumotlari")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

		// Buyurtmani tasdiqlash tugmasini yuborish
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Tasdiqlash", fmt.Sprintf("book_%s_%s_%s", barberName, orderDate, orderTime)),
				tgbotapi.NewInlineKeyboardButtonData("Ortga qaytish", "back"),
			),
		)

		confirmationMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Siz %s kuni %s vaqtida buyurtma qilishni xohlaysizmi?", orderDate, orderTime))
		confirmationMsg.ReplyMarkup = &keyboard
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Tasdiqlash xabarini yuborishda xatolik: %v", err)
			return
		}

		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()
	} else if strings.HasPrefix(callbackData, "book_") {
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Buyurtma qilish uchun noto'g'ri callback ma'lumotlari")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

		// Foydalanuvchida mavjud bo'lgan in_process statusli buyurtmani tekshirish
		hasInProcess, err := storage.HasInProcessOrder(int64(update.CallbackQuery.From.ID))
		if err != nil {
			log.Printf("Buyurtmalarni tekshirishda xatolik: %v", err)
			return
		}
		if hasInProcess {
			confirmationMsg := tgbotapi.NewMessage(chatID, "Sizda hali bajarilmagan buyurtma mavjud. Yangi buyurtma qilishdan oldin uni yakunlang.")
			botInstance.Send(confirmationMsg)
			return
		}

		// Buyurtmani saqlash
		order := models.Order{
			BarberName: barberName,
			UserID:     update.CallbackQuery.From.ID,
			OrderDate:  orderDate,
			OrderTime:  orderTime,
			Status:     "in_process",
		}

		if err := storage.SaveOrder(order); err != nil {
			log.Printf("Buyurtmani saqlashda xatolik: %v", err)
			return
		}

		userStates.RLock()
		userStates.RUnlock()
		deleteMessage := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
		botInstance.Send(deleteMessage)

		confirmationMsg := tgbotapi.NewMessage(chatID, "Buyurtma muvaffaqiyatli saqlandi!")
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Tasdiqlash xabarini yuborishda xatolik: %v", err)
		}

		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()

		sendOrderDetailsToBarber(order, botInstance)
	} else if callbackData == "back" {
		userStates.RLock()
		lastMessageID := userStates.m[chatID]
		userStates.RUnlock()
		barberName := strings.Split(callbackData, "_")[1]
		orderDate := strings.Split(callbackData, "_")[2]
		SelectOrder(chatID, botInstance, barberName, orderDate, update, lastMessageID)
	}
}

func sendOrderDetailsToBarber(order models.Order, botInstance *tgbotapi.BotAPI) {
	barbers, err := storage.GetBarbers()
	if err != nil {
		log.Printf("Sartaroshlarni olishda xatolik: %v", err)
		return
	}

	var barberChatID int64
	for _, barber := range barbers {
		if barber.Name == order.BarberName {
			barberChatID = barber.ID
			break
		}
	}

	if barberChatID == 0 {
		log.Printf("Sartarosh chat ID topilmadi: %s", order.BarberName)
		return
	}

	user := storage.GetUserFromDB(int64(order.UserID))

	message := fmt.Sprintf("Yangi buyurtma:\nSartarosh: %s\nSana: %s\nVaqt: %s\nFoydalanuvchi telefon raqami: %s", order.BarberName, order.OrderDate, order.OrderTime, user.Phone)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Buyurtmani bekor qilish", fmt.Sprintf("cancel_%s_%s_%s", order.BarberName, order.OrderDate, order.OrderTime)),
			tgbotapi.NewInlineKeyboardButtonData("Buyurtma bajarildi", fmt.Sprintf("done_%s_%s_%s", order.BarberName, order.OrderDate, order.OrderTime)),
		),
	)

	barberMessage := tgbotapi.NewMessage(barberChatID, message)
	barberMessage.ReplyMarkup = keyboard

	_, err = botInstance.Send(barberMessage)
	if err != nil {
		log.Printf("Buyurtma ma'lumotlarini sartaroshga yuborishda xatolik: %v", err)
	}
}


func HandleCancelOrder(callback *tgbotapi.CallbackQuery, update tgbotapi.Update) {
	data := strings.Split(callback.Data, "_")
	if len(data) < 4 {
		log.Println("Bekor qilish uchun noto'g'ri callback ma'lumotlari")
		return
	}
	barberName := data[1]
	orderDate := data[2]
	orderTime := data[3]

	userID := storage.GetUserIDByOrderDetails(barberName, orderDate, orderTime)

	err := storage.DeleteOrder(barberName, orderDate, orderTime)
	if err != nil {
		log.Printf("Buyurtmani o'chirishda xatolik: %v", err)
		botInstance := config.GetBot()
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Buyurtmani o'chirishda xatolik yuz berdi.")
		botInstance.Send(msg)
		return
	}

	botInstance := config.GetBot()
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Buyurtma muvaffaqiyatli bekor qilindi.")
	botInstance.Send(msg)

	// Foydalanuvchiga xabar yuborish
	if userID != 0 {
		userMsg := tgbotapi.NewMessage(userID, "Sizning buyurtmangiz bekor qilindi.")
		botInstance.Send(userMsg)
	}
}

func HandleCompleteOrder(callback *tgbotapi.CallbackQuery, update tgbotapi.Update) {
	data := strings.Split(callback.Data, "_")
	if len(data) < 4 {
		log.Println("Tasdiqlash uchun noto'g'ri callback ma'lumotlari")
		return
	}
	barberName := data[1]
	orderDate := data[2]
	orderTime := data[3]

	err := storage.CompleteOrder(barberName, orderDate, orderTime)
	if err != nil {
		log.Printf("Buyurtmani yakunlashda xatolik: %v", err)
		botInstance := config.GetBot()
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Buyurtmani yakunlashda xatolik yuz berdi.")
		botInstance.Send(msg)
		return
	}

	botInstance := config.GetBot()
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Buyurtma muvaffaqiyatli yakunlandi.")
	botInstance.Send(msg)

	// Foydalanuvchiga xabar yuborish
	userID := storage.GetUserIDByOrderDetails(barberName, orderDate, orderTime)
	if userID != 0 {
		userMsg := tgbotapi.NewMessage(userID, "Sizning buyurtmangiz bajarildi.")
		botInstance.Send(userMsg)
	}
}

