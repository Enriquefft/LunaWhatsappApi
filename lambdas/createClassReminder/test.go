// package main
//
// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// 	"strings"
//
// 	// "github.com/aws/aws-lambda-go/lambda"
// 	openai "github.com/sashabaranov/go-openai"
// )
//
// // No parameters
// type Schedule struct {
// 	UserName    string `json:"userName"`
// 	UserPhone   string `json:"userPhone"`
// 	EventName   string `json:"name"`
// 	Description string `json:"description"`
// 	StartTime   string `json:"startTime"`
// 	EndTime     string `json:"endTime"`
// }
//
// func getTimeDifference(startTime, endTime string) (string, string) {
// 	startParts := strings.Split(startTime, ":")
// 	endParts := strings.Split(endTime, ":")
//
// 	startHour, _ := strconv.Atoi(startParts[0])
// 	startMinute, _ := strconv.Atoi(startParts[1])
// 	endHour, _ := strconv.Atoi(endParts[0])
// 	endMinute, _ := strconv.Atoi(endParts[1])
//
// 	startMinutes := startHour*60 + startMinute
// 	endMinutes := endHour*60 + endMinute
//
// 	differenceMinutes := endMinutes - startMinutes
// 	hours := fmt.Sprint(differenceMinutes / 60)
// 	minutes := fmt.Sprint(differenceMinutes % 60)
//
// 	return hours, minutes
// }
//
// func getPrompt(schedule *Schedule) string {
// 	hour, minutes := getTimeDifference(schedule.StartTime, schedule.EndTime)
// 	prompt := `Por favor, crea un recordatorio para ` + schedule.UserName + ` para la clase de ` +
// 		schedule.EventName + `. La clase comenzará pronto y tendrá una duración de ` + hour + ` horas`
//
// 	if minutes != "0" {
// 		prompt += ` y ` + minutes + ` minutos`
// 	}
//
// 	prompt += `. Aquí tienes una breve descripción de la clase: ` + schedule.Description + `. ` +
// 		`Asegúrate de que el recordatorio tenga un tono amigable y sea conciso.`
//
// 	fmt.Println(prompt)
//
// 	return prompt
// }
//
// func HandleRequest(ctx context.Context, data Schedule) (string, error) {
// 	openai_api_key := os.Getenv("OPENAI_API_KEY")
// 	openai_client := openai.NewClient(openai_api_key)
//
// 	resp, response_err := openai_client.CreateChatCompletion(
// 		context.Background(),
// 		openai.ChatCompletionRequest{
// 			Model: openai.GPT3Dot5Turbo,
// 			Messages: []openai.ChatCompletionMessage{
// 				{
// 					Role:    openai.ChatMessageRoleSystem,
// 					Content: getPrompt(&data),
// 				},
// 			},
// 		},
// 	)
//
// 	if response_err != nil {
// 		log.Printf("ChatCompletion error: %v\n", response_err)
// 		return "", response_err
// 	}
//
// 	return resp.Choices[0].Message.Content, nil
// }
//
// func main() {
// 	test := Schedule{
// 		EventName: "Mate",
// 		StartTime: "12:00",
// 		EndTime:   "13:00",
// 		UserName:  "enrique",
// 		UserPhone: "984724707",
// 	}
//
// 	res, err := HandleRequest(context.TODO(), test)
// 	fmt.Println(res)
// 	if err != nil {
// 		fmt.Println("Got an error")
// 		fmt.Println(err)
// 	}
// }
