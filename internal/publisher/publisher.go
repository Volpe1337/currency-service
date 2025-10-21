package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"currency_service/proto"

	natsproto "google.golang.org/protobuf/proto"

	"github.com/nats-io/nats.go"
	"github.com/valyala/fasthttp"
)

// CBRResponse структура для парсинга ответа от ЦБ РФ
type CBRResponse struct {
	Valute map[string]CurrencyInfo `json:"Valute"`
}

// CurrencyInfo структура для информации о валюте от ЦБ РФ
type CurrencyInfo struct {
	ID       string  `json:"ID"`
	NumCode  string  `json:"NumCode"`
	CharCode string  `json:"CharCode"`
	Nominal  int     `json:"Nominal"`
	Name     string  `json:"Name"`
	Value    float64 `json:"Value"`
	Previous float64 `json:"Previous"`
}

// Start запускает отправитель данных о валютах
func Start() error {
	// Подключаемся к серверу Nats
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Nats: %w", err)
	}
	defer nc.Close()

	fmt.Println("Отправитель запущен. Опрос ЦБ РФ каждые 30 секунд...")

	// Бесконечный цикл с опросом ЦБ
	for {
		if err := fetchAndSendCurrencies(nc); err != nil {
			log.Printf("Ошибка получения/отправки данных: %v", err)
		}

		// Ждем 30 секунд перед следующим запросом
		time.Sleep(30 * time.Second)
	}
}

// fetchAndSendCurrencies получает данные от ЦБ и отправляет через Nats
func fetchAndSendCurrencies(nc *nats.Conn) error {
	fmt.Println("Запрашиваем данные у ЦБ РФ...")

	// Создаем HTTP клиент fasthttp
	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	// Настраиваем запрос
	req.SetRequestURI("https://www.cbr-xml-daily.ru/daily_json.js")
	req.Header.SetMethod("GET")

	// Выполняем запрос
	if err := client.Do(req, resp); err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	// Проверяем статус код
	if statusCode := resp.StatusCode(); statusCode != 200 {
		return fmt.Errorf("неверный статус код: %d", statusCode)
	}

	// Парсим JSON ответ
	var cbrResponse CBRResponse
	if err := json.Unmarshal(resp.Body(), &cbrResponse); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	// Создаем список интересующих нас валют
	targetCurrencies := []string{"USD", "EUR", "GBP", "CNY", "JPY"}

	// Создаем protobuf сообщение
	var currencies []*proto.Currency

	for _, code := range targetCurrencies {
		if currencyInfo, exists := cbrResponse.Valute[code]; exists {
			currencies = append(currencies, &proto.Currency{
				Name:  currencyInfo.Name,
				Code:  currencyInfo.CharCode,
				Value: currencyInfo.Value,
			})
			fmt.Printf("  Добавлена валюта: %s - %.4f руб.\n", currencyInfo.CharCode, currencyInfo.Value)
		}
	}

	if len(currencies) == 0 {
		return fmt.Errorf("не найдено ни одной валюты")
	}

	currencyResponse := &proto.CurrencyResponse{
		Currencies: currencies,
	}

	// Кодируем в protobuf
	data, err := natsproto.Marshal(currencyResponse)
	if err != nil {
		return fmt.Errorf("ошибка кодирования protobuf: %w", err)
	}

	// Отправляем через Nats
	if err := nc.Publish("currencies", data); err != nil {
		return fmt.Errorf("ошибка отправки через Nats: %w", err)
	}

	fmt.Printf("Успешно отправлено %d валют через Nats\n", len(currencies))
	return nil
}
