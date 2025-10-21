package main

import (
	"currency_service/internal/subscriber"
	"log"
)

func main() {
	if err := subscriber.Start(); err != nil {
		log.Fatal("Ошибка запуска подписчика:", err)
	}
}
