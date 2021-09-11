package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Covid19 struct {
	TxnDate                string `json:"txn_date"`
	NewCase                int    `json:"new_case"`
	TotalCase              int    `json:"total_case"`
	NewCaseExcludeabroad   int    `json:"new_case_excludeabroad"`
	TotalCaseExcludeabroad int    `json:"total_case_excludeabroad"`
	NewDeath               int    `json:"new_death"`
	TotalDeath             int    `json:"total_death"`
	NewRecovered           int    `json:"new_recovered"`
	TotalRecovered         int    `json:"total_recovered"`
	UpdateDate             string `json:"update_date"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()
	app.Get("/", get)

	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}

func getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func get(c *fiber.Ctx) error {
	data, err := getRedis("data")
	if err != nil {
		c.JSON(err)
	}

	if data == nil {
		data, err := fetchData()
		if err != nil {
			c.JSON(err)
		}

		if err := Set("data", data); err != nil {
			c.JSON(err)
		}

		fmt.Println("set data")

		return c.Status(fiber.StatusOK).JSON(data)
	}

	return c.Status(fiber.StatusOK).JSON(data)

	// return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Hello, World"})
}

func Set(key string, value interface{}) error {
	rdb := getClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	json, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := rdb.Set(ctx, key, json, 10*time.Second).Err(); err != nil {
		return err
	}

	return nil
}

func getRedis(key string) ([]Covid19, error) {
	rdb := getClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var covid []Covid19
	if err := json.Unmarshal([]byte(val), &covid); err != nil {
		panic(err)
	}

	return covid, nil
}

func fetchData() ([]Covid19, error) {
	uri := os.Getenv("URI")
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []Covid19

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	// for _, item := range data {
	// 	return &item, nil
	// }

	return data, nil

}
