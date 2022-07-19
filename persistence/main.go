package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/nats-io/stan.go"
	"log"
	"net/http"
	"os"
)

func main() {
	url := configure()
	db, err := pgx.Connect(context.Background(), url)
	CheckError(err)
	defer db.Close(context.Background())
	err = InitCache(db)
	CheckError(err)

	// Подписка на канал в nuts-streaming
	sc, _ := stan.Connect("prod", "sub-1")
	defer sc.Close()
	sc.Subscribe("example", func(msg *stan.Msg) {
		order, err := JsonToObject(string(msg.Data))
		fmt.Println(order)
		if err != nil {
			fmt.Println("Некорректные данные на входе")
		} else {
			err := InsertData(db, order)
			if err != nil {
				fmt.Println("некорректные данные")
			}
		}
	})

	// Запуск веб сервера и прослушивание запросов
	http.HandleFunc("/index", IndexHandler)
	http.HandleFunc("/order", OrderHandler)
	err = http.ListenAndServe("localhost:8080", nil)
	CheckError(err)
}

/*
Конфигурация приложения
P.S.
пробовал без инициализации через файл и просто передавал ссылку типа
urlExample := "postgres://username:password@localhost:5432/database_name"

PostgreSQL выкидывала непонятные ошибки поэтому реализовал через переменные в файле
*/
func configure() string {
	err := godotenv.Load()
	CheckError(err)
	return fmt.Sprintf("%v://%v:%v@%v:%v/%v", os.Getenv("DRIVER"), os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"), os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("DB_NAME"))
}

// CheckError Обработка ошибки
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
