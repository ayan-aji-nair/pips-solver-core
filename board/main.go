package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"pips-solver/backend/board/requests"
	"pips-solver/backend/board/solver"
	"time"
)

// TODO Add proper logging
// TODO Read from .env file
// TODO Expose solver endpoint as API

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Errorf("Error loading dotenv file")
	}

	base_url := os.Getenv("BASE_URL")
	token := os.Getenv("TOKEN")
	date := os.Getenv("DATE")

	http_client := &http.Client{
		Timeout: 10 * time.Second,
	}

	c, err := requests.NewClient(base_url, token, http_client)
	if err != nil {
		fmt.Errorf("TestGetPuzzlesRequest failed, client creation error")
	}

	ctx := context.Background()
	out, err := c.GetPuzzles(ctx, date)

	if err != nil {
		fmt.Errorf("TestGetPuzzlesRequest failed, could not get data")
	}

	jsonBytes, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		fmt.Errorf("TestGetPuzzlesRequest failed, error marshalling")
	}

	fmt.Println(string(jsonBytes))

	model, err := solver.NewILPModel(out.Data.Hard)
	if err != nil {
		fmt.Errorf("NewILPModel failed: %v", err)
	}

	chosen, err := model.Solve()
	if err != nil {
		fmt.Errorf("Solve failed: %v", err)
	}

	jsonBytes, err = json.MarshalIndent(chosen, "", "    ")
	if err != nil {
		fmt.Errorf("TestGetPuzzlesRequest failed, error marshalling")
	}

	fmt.Println(string(jsonBytes))
}
