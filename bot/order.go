// bot/order.go

package bot

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// SelectBarber sartaroshni rasm bilan jo'natadi
func SelectBarber(chatID int64, botInstance *tgbotapi.BotAPI) {
	// Rasmning yo'lini belgilash
	photoFilePath := "./photo.jpg"
	caption := "Iltimos, sartaroshni tanlang:"

	// Rasm faylini ochish va o'qish
	photoFileBytes, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Rasm faylini o'qishda xatolik: %v", err)
		return
	}

	// Jo'natiladigan rasm uchun yangi rasm konfiguratsiyasi yaratish
	msg := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoFileBytes,
	})

	// Rasm sarlavhasini sozlash
	msg.Caption = caption

	// Ikki qatorlik sartaroshni tanlash tugmalarini o'z ichiga olgan to'plam yaratish
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// Birinchi qator (sartarosh tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Baxtiyor", "select_date_Baxtiyor"),
			tgbotapi.NewInlineKeyboardButtonData("Ali", "select_date_Ali"),
		),
		// Ikkinchi qator (sartarosh tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Aziz", "select_date_Aziz"),
			tgbotapi.NewInlineKeyboardButtonData("Obid", "select_date_Obid"),
		),
	)

	// Sartarosh tanlash tugmalari bilan sartarosh tanlash to'plamini bog'lash
	msg.ReplyMarkup = &keyboard

	// Rasm xabarni jo'natish
	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Rasmni jo'natishda xatolik: %v", err)
	}

	// Logda sartarosh tanlanganda chiqariladigan xabar
	log.Println("Sartaroshni tanlash tugmalarini jo'natish uchun botdan so'roq jo'natildi")
}

// SelectDate foydalanuvchiga sartarosh tanlagandan so'ng sanani tanlashni so'rash
func SelectDate(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, update tgbotapi.Update) {
	// Agar sartarosh nomi bo'sh bo'lsa, sanani tanlashni so'ramaslik
	if barberName == "" {
		log.Println("Sartarosh tanlanmagan. Sanani tanlash o'tkaziladi.")
		return
	}

	fmt.Println("Sanaga kirdi")

	// Bugungi sanani olish
	today := time.Now()
	// Ertangi sanani olish
	tomorrow := today.AddDate(0, 0, 1)
	// Ertagina kechasan sanani olish
	dayAfterTomorrow := today.AddDate(0, 0, 2)

	// Sanani tanlash tugmalari bilan ichki sartaroshni yaratish
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// Birinchi qator (sanani tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(today.Format("02.01.2006"), fmt.Sprintf("datte_%s", today.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(tomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s", tomorrow.Format("2006-01-02"))),
		),
		// Ikkinchi qator (sanani tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dayAfterTomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s", dayAfterTomorrow.Format("2006-01-02"))),
		),
	)

	// Sanani tanlash tugmalari bilan sanani jo'natish
	dateSelectionMsg := tgbotapi.NewMessage(chatID, "Sana tanlang (kun.oy.yil):")
	dateSelectionMsg.ReplyMarkup = &keyboard
	if _, err := botInstance.Send(dateSelectionMsg); err != nil {
		log.Printf("Sanani tanlash klaviaturasini jo'natishda xatolik: %v", err)
		return
	}
}

func SelectOrder(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, update tgbotapi.Update) {
	callbackData := update.CallbackQuery.Data

	fmt.Println("Orderga kirdi")

	fmt.Println("Sartarosh: ", barberName)

	if strings.Contains(strings.ToLower(callbackData), "datte_") {
		fmt.Println("SANA:    ", callbackData)
	}
}
