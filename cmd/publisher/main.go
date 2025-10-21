package main

import (
	"currency_service/internal/publisher"
	"log"
)

func main() {
	if err := publisher.Start(); err != nil {
		log.Fatal("Ошибка запуска отправителя:", err)
	}
}
