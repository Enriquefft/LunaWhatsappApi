package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	start_lambda "github.com/aws/aws-lambda-go/lambda"
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

type Response struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

func sendMessage(ctx context.Context, response *Response) error {
	cfg, aws_config_err := config.LoadDefaultConfig(context.TODO())
	if aws_config_err != nil {
		log.Println("Error loading AWS configuration:", aws_config_err)
		return aws_config_err
	}

	svc := lambda.NewFromConfig(cfg)

	payload, json_err := json.Marshal(response)

	if json_err != nil {
		return json_err
	}

	_, invoke_error := svc.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String("sendMessage"),
		Payload:      payload,
	})
	return invoke_error
}

func getTimeDifference(startTime, endTime string) (string, string) {
	startParts := strings.Split(startTime, ":")
	endParts := strings.Split(endTime, ":")

	startHour, _ := strconv.Atoi(startParts[0])
	startMinute, _ := strconv.Atoi(startParts[1])
	endHour, _ := strconv.Atoi(endParts[0])
	endMinute, _ := strconv.Atoi(endParts[1])

	startMinutes := startHour*60 + startMinute
	endMinutes := endHour*60 + endMinute

	differenceMinutes := endMinutes - startMinutes
	hours := fmt.Sprint(differenceMinutes / 60)
	minutes := fmt.Sprint(differenceMinutes % 60)

	return hours, minutes
}

// Categoria clase
func getPrompt(schedule *Schedule) string {
	hour, minutes := getTimeDifference(schedule.StartTime, schedule.EndTime)
	prompt := `Por favor, crea un recordatorio para ` + schedule.UserName + ` para la clase de ` +
		schedule.EventName + `. La clase comenzará pronto y tendrá una duración de ` + hour + ` horas`

	if minutes != "0" {
		prompt += ` y ` + minutes + ` minutos`
	}

	prompt += `. Aquí tienes una breve descripción de la clase: ` + schedule.Description + `. ` +
		`Asegúrate de que el recordatorio tenga un tono amigable y sea conciso.`

	return prompt
}

func HandleRequest(ctx context.Context, data Schedule) (Response, error) {
	openai_api_key := os.Getenv("OPENAI_API_KEY")
	openai_client := openai.NewClient(openai_api_key)

	resp, response_err := openai_client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: getPrompt(&data),
				},
			},
		},
	)

	if response_err != nil {
		log.Printf("ChatCompletion error: %v\n", response_err)
		return Response{}, response_err
	}

	response := Response{Phone: data.UserPhone, Message: resp.Choices[0].Message.Content}
	send_error := sendMessage(ctx, &response)

	if send_error != nil {
		log.Printf("Error sending message: %v\n", send_error)
		return Response{}, send_error
	}

	return response, nil
}

func main() {
	start_lambda.Start(HandleRequest)
}
