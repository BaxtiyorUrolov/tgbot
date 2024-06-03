package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tgbot/bot"
	"tgbot/config"
	"tgbot/models"
	"tgbot/state"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	dbConfig := models.DB{
		Host:     "localhost",
		Port:     5432,
		User:     "godb",
		Password: "0208",
		Name:     "tgbot",
	}
	botToken := "6588290150:AAH4WJpey4hnCVj7Dtr8P-l_9nlssQRdWo0"

	if err := config.Setup(dbConfig, botToken); err != nil {
		log.Fatalf("Konfiguratsiyani sozlashda xatolik: %v", err)
	}

	defer func() {
		if err := config.GetDB().Close(); err != nil {
			log.Printf("Ma'lumotlar omborini yopishda xatolik: %v", err)
		}
	}()

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Fatal("Bot instansiyasini olishda xatolik")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
				time.Sleep(5 * time.Second)
				continue
			}
			for _, update := range updates {
				HandleUpdate(update)
				offset = update.UpdateID + 1
			}
		}
	}
}

func HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		handleMessage(update.Message)
	} else if update.InlineQuery != nil {
		handleInlineQuery(update.InlineQuery)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(update)
	} else {
		log.Printf("Qabul qilingan qo'llab-quvvatlanmaydigan yangilanish turi: %T", update)
	}
}

func handleInlineQuery(inline *tgbotapi.InlineQuery) {
	log.Printf("Inline so'rov qabul qilindi: %s", inline.Query)
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

	if strings.HasPrefix(data, "select_date_") {
		barberName := strings.TrimPrefix(data, "select_date_")

		state.UserStates.Lock()
		state.UserStates.M[chatID] = "select_date"
		state.UserStates.Unlock()

		bot.SelectDate(chatID, botInstance, barberName, callback.Message.MessageID)
	} else if strings.HasPrefix(data, "datte_") {
		dataParts := strings.Split(strings.TrimPrefix(data, "datte_"), "_")
		if len(dataParts) < 2 {
			log.Println("Buyurtma tanlash uchun noto'g'ri callback ma'lumotlari")
			return
		}
		barberName := dataParts[0]
		orderDate := dataParts[1]

		state.UserStates.RLock()
		state.UserStates.RUnlock()

		bot.SelectOrder(chatID, botInstance, barberName, orderDate, update, callback.Message.MessageID)
	} else if strings.HasPrefix(data, "confirm_") || strings.HasPrefix(data, "book_") || data == "back" {
		bot.HandleConfirmation(chatID, botInstance, callback, update)
	} else if strings.HasPrefix(data, "cancel_") {
		bot.HandleCancelOrder(callback, update)
	} else if strings.HasPrefix(data, "done_") {
		bot.HandleCompleteOrder(callback, update)
	} else {
		log.Printf("Noma'lum callback ma'lumotlari: %s", data)
	}
}

func handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text
	botInstance := config.GetBot()

	log.Printf("Received message from chat ID %d: %s", chatID, text)

	state.UserStates.RLock()
	currentState, exists := state.UserStates.M[chatID]
	state.UserStates.RUnlock()

	log.Printf("Current state for chat ID %d: %s", chatID, currentState)

	if text == "/start" {
		handleStartCommand(msg)
	} else if text == "/admin" {
		handleAdminCommand(msg)
	} else if text == "/barber" {
		handleBarberCommand(msg)
	} else if text == "Statistika" {
		bot.HandleAdminStatistics(chatID, botInstance)
	} else if text == "Barber qo'shish" {
		bot.HandleAdminAddBarber(chatID, botInstance)
	} else if text == "Buyurtma qo'shish" {
		name := storage.GetBarber(int(chatID))
		bot.SelectDate(chatID, botInstance, name, 0)
	} else if text == "Barber o'chirish" {
		bot.HandleAdminDeleteBarber(chatID, botInstance)
	} else {
		if !exists {
			handleDefaultMessage(msg)
		} else {
			switch currentState {
			case "adding_barber_id":
				log.Println("Handling state: adding_barber_id")
				bot.HandleAddBarberID(msg)
			case "adding_barber_name":
				log.Println("Handling state: adding_barber_name")
				bot.HandleAddBarberName(msg)
			case "adding_barber_username":
				log.Println("Handling state: adding_barber_username")
				bot.HandleAddBarberUserName(msg)
			case "adding_barber_phone":
				log.Println("Handling state: adding_barber_phone")
				bot.HandleAddBarberPhone(msg)
			case "register_name":
				log.Println("Handling state: register_name")
				bot.HandleRegister(msg)
			case "register_phone":
				log.Println("Handling state: register_phone")
				bot.HandleRegister(msg)
			default:
				log.Println("Handling default message")
				handleDefaultMessage(msg)
			}
		}
	}
}

func handleAdminCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	adminType := storage.BarberType(int(chatID))

	if adminType == "admin" {
		bot.Admin(chatID, botInstance)
	}
}

func handleBarberCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	adminType := storage.BarberType(int(chatID))

	if adminType == "barber" {
		bot.Barber(chatID, botInstance)
	}
}

func handleStartCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	log.Printf("Start komandasi bajarilmoqda, chat ID: %d, user ID: %d", chatID, userID)

	db := config.GetDB()
	if db == nil {
		log.Println("Ma'lumotlar ombori bilan ulanish muammosi")
		return
	}

	log.Println("Foydalanuvchi ma'lumotlar omboriga kiritilmoqda")
	if _, err := db.Exec(`INSERT INTO users (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, userID); err != nil {
		log.Printf("Foydalanuvchini ma'lumotlar omboriga kiritishda xatolik: %v", err)
		return
	}

	log.Println("Foydalanuvchi ma'lumotlari olinmoqda")
	user := storage.GetUserFromDB(int64(userID))
	log.Printf("Olingan foydalanuvchi: %+v", user)

	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instansiyasi bilan muammo")
		return
	}

	if user == nil || user.Name == "" || user.Phone == "" {
		log.Println("Foydalanuvchi ro'yxatdan o'tmagan yoki ismi yoki telefoni bo'sh")
		bot.Register(chatID, botInstance)
		state.UserStates.Lock()
		state.UserStates.M[chatID] = "register_name"
		log.Printf("Foydalanuvchi holati yangilandi: %d -> %s", chatID, "register_name")
		state.UserStates.Unlock()
	} else {
		message := fmt.Sprintf("Assalomu alaykum, %s! Botga xush kelibsiz.", user.Name)
		msgSend := tgbotapi.NewMessage(chatID, message)
		log.Printf("Xabar yuborilmoqda: %s", message)

		_, err := botInstance.Send(msgSend)
		if err != nil {
			log.Printf("Xabar yuborishda xatolik: %v", err)
			return
		}

		bot.SelectBarber(chatID, botInstance)
	}
}

func handleDefaultMessage(msg *tgbotapi.Message) {
	// Har qanday boshqa xabarlarni shu yerda ko'rib chiqish mumkin
}
