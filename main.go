package main

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const ConfigPath = "config"
const TimeToSleep = time.Second * 0
const CoolDown = 60 * 24
const CallingInterval = time.Second * 60
var infoKeeper InfoKeeper
var config Config
var FeaturePattern = regexp.MustCompile("/set\\s(.+)")

const StartText = `Привет! Я бот для определения особенных людей!
Список команд:
/help — вывести текущее сообщение;
/spin — определение особенного человека;
/set — установка особенности;
/stat — показать статистику особенных людей;
/stat — показать топ особенных людей;
/members — показать людей, которые могут быть особенными;
/chat_id — показать id чата;
/join — добавить себя во множество людей, которые могут быть особенными;
/ping — проверить работоспособность бота.
`


func getUserCallName(user *tgbotapi.User) string {
	if user.UserName == "" {
		return fmt.Sprintf("@%v", user.ID)
	}
	return fmt.Sprintf("@%v", user.UserName)
}

func handleStart(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, StartText)
	if _, err := bot.Send(msg); err != nil {
		log.Println("Handle start error: " + err.Error())
	}
}

func handleSpin(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	if !infoKeeper.IsChatJoined(chatId) {
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, "Нельзя выбирать человека из пустого множества.")); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
	}
	info, err := infoKeeper.Read(chatId)
	if err != nil {
		log.Println("Handle spin error: " + err.Error())
		return
	}
	currentTime := time.Now().Unix()
	var userIds []int
	for newUserId := range info.Users {
		userIds = append(userIds, newUserId)
	}

	if int64(info.LastSpinTime) < currentTime-24*60*60 {
		specialHumanId := userIds[rand.Intn(len(userIds))]
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, "Итак, начинаем ультрасложный подсчёт...")); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		time.Sleep(TimeToSleep)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, GenExpression())); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		time.Sleep(TimeToSleep)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, GenInequality())); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		time.Sleep(TimeToSleep)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, GenSqrt())); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		time.Sleep(TimeToSleep)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, "...")); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		time.Sleep(TimeToSleep)
		responseMessage := fmt.Sprintf("Сегодня %v дня — %v! Поздравляем его!", info.Feature, info.Users[specialHumanId].Name)

		specialHuman := info.Users[specialHumanId]
		specialHuman.Count += 1
		info.Users[specialHumanId] = specialHuman

		info.SpecialHuman = specialHumanId
		lastMessage, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage))
		if err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
		_, err = bot.PinChatMessage(tgbotapi.PinChatMessageConfig{
			ChatID:              lastMessage.Chat.ID,
			MessageID:           lastMessage.MessageID,
			DisableNotification: true,
		})
		if err != nil {
			log.Println("Handle spin error: " + err.Error())
		}
		info.LastSpinTime = currentTime
	} else {
		respMessage := fmt.Sprintf(
			"По уже проведённому подсчёту сегодня %v дня — %v.",
			info.Feature,
			info.Users[info.SpecialHuman].Name,
		)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, respMessage)); err != nil {
			log.Println("Handle spin error: " + err.Error())
			return
		}
	}
	if err := infoKeeper.Write(chatId, info); err != nil {
		log.Println("Handle spin error: " + err.Error())
		return
	}
}

func handleJoin(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	user := update.Message.From
	if !infoKeeper.IsChatJoined(chatId) {
		info := Info{
			Users: map[int]User{
				user.ID: {
					Name:     user.FirstName,
					Count:    0,
					Nickname: getUserCallName(user),
				},
			},
			Feature:      "feature",
			LastSpinTime: 0,
			SpecialHuman: user.ID,
		}
		responseMessage := fmt.Sprintf("Теперь пользователь %v может быть особенным!", user.FirstName)
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
			log.Println("Handle join error: " + err.Error())
			return
		}
		if err := infoKeeper.Write(chatId, &info); err != nil {
			log.Println("Handle join error: " + err.Error())
			return
		}
	} else {
		info, err := infoKeeper.Read(chatId)
		if err != nil {
			log.Println("Handle join error: " + err.Error())
			return
		}
		if infoUser, ok := info.Users[user.ID]; ok {
			responseMessage := fmt.Sprintf("Пользователь %v уже есть во множестве людей, которые могут быть особенными!", infoUser.Name)
			if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
				log.Println("Handle join error: " + err.Error())
				return
			}
		} else {
			info.Users[user.ID] = User{
				Name:     user.FirstName,
				Nickname: getUserCallName(user),
				Count:    0,
			}
			if err := infoKeeper.Write(chatId, info); err != nil {
				log.Println("Handle join error: " + err.Error())
				return
			}
			responseMessage := fmt.Sprintf("Теперь пользователь %v может быть особенным!", user.FirstName)
			if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
				log.Println("Handle join error: " + err.Error())
				return
			}
		}
	}
}

func handleMembers(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	if !infoKeeper.IsChatJoined(chatId) {
		responseMessage := fmt.Sprintf("Никто не может быть особенным, потому что никто не присоединился.")
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
			log.Println("Handle members error: " + err.Error())
			return
		}
	} else {
		info, err := infoKeeper.Read(chatId)
		if err != nil {
			log.Println("Handle members error: " + err.Error())
			return
		}
		if _, err := bot.Send(tgbotapi.NewMessage(chatId, buildMembersMessage(info))); err != nil {
			log.Println("Handle members error: " + err.Error())
			return
		}
	}
}

func handleSetFeature(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	message := update.Message.Text
	var responseMessage string
	if !infoKeeper.IsChatJoined(chatId) {
		responseMessage = "Нельзя ставить фичу без людей. Используйте /join."
	} else {
		info, err := infoKeeper.Read(chatId)
		if err != nil {
			log.Println("Handle set feature error: " + err.Error())
			return
		}
		if message == "/set" {
			responseMessage = fmt.Sprintf("Текущая особенность: %v.", info.Feature)
		} else {
			if FeaturePattern.MatchString(message) {
				newFeature := message[len("/set"):]
				responseMessage = fmt.Sprintf("Особенность сменена на %v.", newFeature)
				info.Feature = newFeature
				if err := infoKeeper.Write(chatId, info); err != nil {
					log.Println("Handle set feature error: " + err.Error())
					return
				}
			}
		}
	}
	if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
		log.Println("Handle set feature error: " + err.Error())
		return
	}
}

func makeStat(bot *tgbotapi.BotAPI, update *tgbotapi.Update, sorted bool) {
	chatId := update.Message.Chat.ID
	var responseMessage string
	if !infoKeeper.IsChatJoined(chatId) {
		responseMessage = "Нельзя смотреть статистику без присоединившихся людей. Используйте /join."
	} else {
		info, err := infoKeeper.Read(chatId)
		if err != nil {
			log.Println("Handle make stat error: " + err.Error())
			return
		}
		responseMessage = buildStatMessage(info, sorted)
	}
	message := tgbotapi.NewMessage(chatId, responseMessage)
	message.ParseMode = "Markdown"
	if _, err := bot.Send(message); err != nil {
		log.Println("Handle make stat error: " + err.Error())
		return
	}
}

func handleStat(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	makeStat(bot, update, false)
}

func handleTop(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	makeStat(bot, update, true)
}

func handleResetDay(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message.From.ID != config.AdminId {
		return
	}
	chatId := update.Message.Chat.ID
	info, err := infoKeeper.Read(chatId)
	if err != nil {
		log.Println("Handle reset error: " + err.Error())
		return
	}
	info.LastSpinTime = 0
	if err := infoKeeper.Write(chatId, info); err != nil {
		log.Println("Handle reset error: " + err.Error())
		return
	}
	if _, err := bot.Send(tgbotapi.NewMessage(chatId, "День был сброшен.")); err != nil {
		log.Println("Handle reset error: " + err.Error())
		return
	}
}

func handlePing(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "pong")); err != nil {
		log.Println("Handle reset error: " + err.Error())
		return
	}
}

func handleChatId(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(update.Message.Chat.ID))); err != nil {
		log.Println("Handle chat id error: " + err.Error())
		return
	}
}

func callSpin(bot *tgbotapi.BotAPI) {
	for {
		chatIds, err := infoKeeper.GetChats()
		if err != nil {
			log.Println("Handle call spin error: " + err.Error())
		}
		for _, chatId := range chatIds {
			info, err := infoKeeper.Read(chatId)
			if err != nil {
				log.Println("Handle call spin error: " + err.Error())
				return
			}
			timeDiff := (time.Now().Unix() - info.LastSpinTime) / 60
			if timeDiff < CoolDown || len(info.Users) == 0 {
				continue
			}
			if rand.Float32() < float32(1) / (24 * 30) {
				var userIds []int
				for newUserId := range info.Users {
					userIds = append(userIds, newUserId)
				}
				callPerson := userIds[rand.Intn(len(userIds))]
				responseMessage := fmt.Sprintf("Не хочешь вызвать /spin, %v?", info.Users[callPerson].Nickname)
				if _, err := bot.Send(tgbotapi.NewMessage(chatId, responseMessage)); err != nil {
					log.Println("Handle call spin error: " + err.Error())
					return
				}
			}
		}
		time.Sleep(CallingInterval)
	}
}

var handlers = map[string]interface{} {
	"/start": handleStart,
	"/help": handleStart,
	"/spin": handleSpin,
	"/join": handleJoin,
	"/members": handleMembers,
	"/set": handleSetFeature,
	"/stat": handleStat,
	"/top": handleTop,
	"/reset": handleResetDay,
	"/ping": handlePing,
	"/chat_id": handleChatId,
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	configPtr, err := ParseConfig(ConfigPath)
	config = *configPtr
	if err != nil {
		panic(err)
	}
	infoKeeper.Init(config.StoragePath)

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates, err := bot.GetUpdatesChan(updateConfig)

	currentTime := fmt.Sprintf("Start! Time: %v.", time.Now().Format(time.RFC3339))
	startMessage := tgbotapi.NewMessage(int64(config.AdminId), currentTime)
	if _, err := bot.Send(startMessage); err != nil {
		log.Println("Starting message error: " + err.Error())
		return
	}

	go callSpin(bot)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		command := strings.Split(update.Message.Text, " ")[0]
		if handler, ok := handlers[command]; ok {
			handler.(func(*tgbotapi.BotAPI, *tgbotapi.Update))(bot, &update)
		}
	}
}
