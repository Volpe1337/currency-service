package main

import (
	"log"

	"github.com/Volpe1337/currency-service/subscriber/internal"
)

func main() {
	if err := internal.Start(); err != nil {
		log.Fatal("Ошибка запуска подписчика:", err)
	}
}
