package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"encoding/json"
	"net/http"
	"strings"
	"io/ioutil"
	"fmt"
	"time"
	"io"
)

type Config struct {
	TelegramBotToken string
	ServiceUrl string
}

type ParselInfoObject struct {
	List []struct {
		TrackingItem struct {
			СommonStatus string `json:"commonStatus"`
			TrackingHistoryItemList []struct {
				Date 		string `json:"date"`
				Index		string `json:"index"`
				CityName	string `json:"cityName"`
				Description	string `json:"description"`
				HumanStatus 	string `json:"humanStatus"`
			} `json:"trackingHistoryItemList"`
		} `json:"trackingItem"`
	} `json:"list"`
}

func main()  {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}
	// В канал updates получаем все новые сообщения
	for update := range updates {

		parcelId := update.Message.Text
		url := configuration.ServiceUrl + parcelId
		data := httpDo(url,"","GET")

		var object ParselInfoObject
		json.Unmarshal(data, &object)

		if len(object.List) > 0 {

			history := object.List[0].TrackingItem.TrackingHistoryItemList

			txt := ""

			for _, item := range history {
				txt += fmt.Sprintf("*%s*\n", item.HumanStatus);

				place := ""

				if item.Index != "" {
					place = fmt.Sprintf("%s, ", item.Index)
				}

				if item.CityName != "" {
					place += item.CityName
				} else if item.Description != "" {
					place += item.Description
				}

				datetime, _ := time.Parse(time.RFC3339, item.Date)

				txt = txt + fmt.Sprintf(
					"%d.%02d.%02d, %02d:%02d %s\n\n",
					datetime.Day(),
					datetime.Month(),
					datetime.Year(),
					datetime.Hour(),
					datetime.Minute(),
					place)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			msg.ParseMode = "markdown"
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправление не найдено")
			bot.Send(msg)
		}
	}
}

func httpDo(url, postData, method string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(method,url,strings.NewReader(postData))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Set("Content-Type", "application/json'")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	out, err := os.Create("response.txt")
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()
	io.Copy(out, resp.Body)

	return body;
}