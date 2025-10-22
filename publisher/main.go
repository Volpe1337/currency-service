package main

import (
	"github.com/Volpe1337/currency-service/publisher/internal"
	"log"
)

func main() {
	if err := internal.Start(); err != nil {
		log.Fatal("Ошибка запуска отправителя:", err)
	}
}
