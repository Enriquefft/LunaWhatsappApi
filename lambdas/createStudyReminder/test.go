package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// "github.com/aws/aws-lambda-go/lambda"
	openai "github.com/sashabaranov/go-openai"
)

// No parameters
type Schedule struct {
	UserName    string `json:"userName"`
	UserPhone   string `json:"userPhone"`
	EventName   string `json:"name"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
}

func HandleRequest(ctx context.Context, data Schedule) (string, error) {
	openai_api_key := os.Getenv("OPENAI_API_KEY")
	openai_client := openai.NewClient(openai_api_key)

	resp, response_err := openai_client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	if response_err != nil {
		log.Printf("ChatCompletion error: %v\n", response_err)
		return "", response_err
	}

	log.Println(resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}

func main() {
	test := Schedule{
		StartTime: "12:00",
		EndTime:   "13:00",
		UserName:  "enrique",
		UserPhone: "984724707",
	}

	res, err := HandleRequest(context.TODO(), test)
	fmt.Println(res)
	if err != nil {
		fmt.Println("Got an error")
		fmt.Println(err)
	}
}
