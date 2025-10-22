package internal

import (
	"fmt"
	"log"

	"github.com/Volpe1337/currency-service/subscriber/models" // Наш пакет

	natsproto "google.golang.org/protobuf/proto" // Пакет protobuf с псевдонимом

	"github.com/nats-io/nats.go"
)

// Start запускает подписчика на получение сообщений о валютах
func Start() error {
	// Подключаемся к серверу Nats
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Nats: %w", err)
	}
	defer nc.Close()

	fmt.Println("Подписчик запущен. Ожидание данных о валютах...")

	// Подписываемся на канал "currencies"
	_, err = nc.Subscribe("currencies", func(msg *nats.Msg) {
		// Распаковываем полученное сообщение
		var currencyResponse models.CurrencyResponse
		err := natsproto.Unmarshal(msg.Data, &currencyResponse) // Используем псевдоним
		if err != nil {
			log.Println("Ошибка распаковки сообщения:", err)
			return
		}

		// Выводим полученные данные в консоль
		fmt.Println("Получены обновления курсов валют:")
		for _, currency := range currencyResponse.Currencies {
			fmt.Printf("  %s (%s): %.4f руб.\n", currency.Name, currency.Code, currency.Value)
		}
		fmt.Println("---")
	})

	if err != nil {
		return fmt.Errorf("ошибка подписки: %w", err)
	}

	// Бесконечно ждем сообщений
	select {}
}
