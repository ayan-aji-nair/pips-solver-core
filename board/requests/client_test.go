package requests_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"pips-solver/backend/board/requests"
	"testing"
	"time"
)

func TestGetPuzzlesRequest(t *testing.T) {
	err := godotenv.Load("../../.env")
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
		t.Errorf("TestGetPuzzlesRequest failed, client creation error")
	}

	ctx := context.Background()
	out, err := c.GetPuzzles(ctx, date)

	if err != nil {
		t.Errorf("TestGetPuzzlesRequest failed, could not get data")
	}

	jsonBytes, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		t.Errorf("TestGetPuzzlesRequest failed, error marshalling")
	}

	// 3. Convert the byte slice directly to a string and print
	fmt.Println(string(jsonBytes))
}
