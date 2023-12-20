package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
)

type Schedule struct {
	UserName    string `json:"userName"`
	UserPhone   string `json:"userPhone"`
	Category    string `json:"category"`
	EventName   string `json:"name"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	WeekDays    []int  `json:"weekDays"`
}

type Response struct {
	IsBase64Encoded bool   `json:"isBase64Encoded"`
	StatusCode      int    `json:"statusCode"`
	Body            string `json:"body"`
}

func getSchedulerClient() (*scheduler.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	scheduler_client := scheduler.NewFromConfig(cfg)

	return scheduler_client, err
}

func formatWeekDays(weekDays []int) string {
	// aws Scheduler week starts on Sunday
	weekDays = incrementWeekDays(weekDays)

	log.Println("formatInput: ", weekDays)
	var builder strings.Builder
	for i, v := range weekDays {
		if i != 0 {
			builder.WriteString(",")
		}
		builder.WriteString(strconv.Itoa(v))
	}

	ans := builder.String()
	log.Println("Week days: ", ans)
	return ans
}

func requiredFields(data *Schedule, extra_required bool) error {
	if extra_required && data.EventName == "" {
		return fmt.Errorf("Event name is required")
	}

	if data.UserName == "" {
		return fmt.Errorf("User name is required")
	}

	if data.UserPhone == "" {
		return fmt.Errorf("User phone is required")
	}

	if data.StartTime == "" {
		return fmt.Errorf("Start time is required")
	}

	if data.EndTime == "" {
		return fmt.Errorf("End time is required")
	}

	if len(data.WeekDays) == 0 {
		return fmt.Errorf("Week days are required")
	}

	return nil
}

func validateInput(data *Schedule) error {
	if data.Category == "" {
		return fmt.Errorf("Category is required")
	}

	// If number of categories increases, this should be changed to a map
	if data.Category != "sleep" && data.Category != "class" && data.Category != "study" {
		return fmt.Errorf("Invalid category")
	}

	// sleep doesn't needs certain fields
	extra_required := data.Category != "sleep"
	err := requiredFields(data, extra_required)
	if err != nil {
		return err
	}

	// Further validation

	// Valid phone number
	if regexp.MustCompile(`^[0-9]{9}$`).MatchString(data.UserPhone) {
		data.UserPhone = "51" + data.UserPhone
	}
	if !regexp.MustCompile(`^[0-9]{11}$`).MatchString(data.UserPhone) {
		return fmt.Errorf("Invalid phone number")
	}

	// Valid time (HH:MM) max 23:59
	if !regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`).MatchString(data.StartTime) {
		return fmt.Errorf("Invalid start time")
	}
	if !regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`).MatchString(data.EndTime) {
		return fmt.Errorf("Invalid end time")
	}

	// Valid week days unique numbers from 1 to 7
	for _, day := range data.WeekDays {
		if day < 1 || day > 7 {
			return fmt.Errorf("Invalid week day")
		}
	}

	return nil
}

func incrementWeekDays(weekDays []int) []int {
	for i, v := range weekDays {
		weekDays[i] = v%7 + 1
	}
	return weekDays
}

func errorResponse(err error) Response {
	return Response{
		IsBase64Encoded: false,
		StatusCode:      400,
		Body:            err.Error(),
	}
}

func HandleRequest(ctx context.Context, data Schedule) (Response, error) {
	// Debugging
	log.Println("Received event: ", data)
	log.Println("Category: ", data.Category)

	// Validate input
	validate_err := validateInput(&data)
	if validate_err != nil {
		log.Println("Got an error validating the input")
		log.Println(validate_err)
		return errorResponse(validate_err), validate_err
	}

	// Scheduler client
	scheduler_client, scheduler_err := getSchedulerClient()
	if scheduler_err != nil {
		log.Println("Got an error retrieving the configuration")
		log.Println(scheduler_err)
		return errorResponse(scheduler_err), scheduler_err
	}

	// Dont need to send the category or the week days
	dataPayload := Schedule{
		UserName:    data.UserName,
		UserPhone:   data.UserPhone,
		EventName:   data.EventName,
		Description: data.Description,
		StartTime:   data.StartTime,
		EndTime:     data.EndTime,
	}

	// If category is "sleep", create two schedules
	if data.Category == "sleep" {
		// Create first schedule at start time
		err := createSchedule(ctx, scheduler_client, dataPayload, data.WeekDays, data.StartTime, "sleep")
		if err != nil {
			return errorResponse(err), err
		}

		// Change day to the next day
		dataPayload.WeekDays = incrementWeekDays(data.WeekDays)

		// Create second schedule at end time
		err = createSchedule(ctx, scheduler_client, dataPayload, data.WeekDays, data.EndTime, "wakeUp")
		if err != nil {
			return errorResponse(err), err
		}
	} else {

		// For other categories, create a single schedule
		err := createSchedule(ctx, scheduler_client, dataPayload, data.WeekDays, data.StartTime, data.Category)
		if err != nil {
			return errorResponse(err), err
		}
	}

	return Response{
		IsBase64Encoded: false,
		StatusCode:      200,
		Body:            "Success",
	}, nil
}

func createSchedule(ctx context.Context, scheduler_client *scheduler.Client, data Schedule, weekDays []int, time string, lambdaCategory string) error {
	ScheduleDescription := fmt.Sprintf("Sheduled event (%s) for %s", data.Category, data.UserName)

	// Target Lambda
	LambdaArn := new(string)

	switch lambdaCategory {
	case "class":
		LambdaArn = aws.String("arn:aws:lambda:us-east-1:354767308165:function:CreateClassReminder")
		log.Println("Assigning class lambda")
	case "sleep":
		LambdaArn = aws.String("arn:aws:lambda:us-east-1:354767308165:function:CreateSleepReminder")
		log.Println("Assigning sleep lambda")
	case "wakeUp":
		LambdaArn = aws.String("arn:aws:lambda:us-east-1:354767308165:function:CreateSleepReminder")
		log.Println("Assigning wakeUp lambda")
	case "study":
		return fmt.Errorf("Study lambda not implemented yet")
	}

	payload, json_err := json.Marshal(data)
	if json_err != nil {
		log.Println("Got an error marshalling the payload")
		return json_err
	}

	ScheduleTarget := &types.Target{
		Arn:     LambdaArn,
		RoleArn: aws.String("arn:aws:iam::354767308165:role/service-role/CreateScheduleRule-role-x8hzay9j"),
		Input:   aws.String(string(payload)),
	}

	week_days := formatWeekDays(weekDays)

	start_time_split := strings.Split(data.StartTime, ":")
	start_hour := start_time_split[0]
	start_minute := start_time_split[1]

	ScheduleExpression := fmt.Sprintf("cron(%s %s ? * %s *)", start_minute, start_hour, week_days)
	log.Println("Schedule Expression: ", ScheduleExpression)

	// Unique Identifier
	ScheduleName := aws.String(strings.ReplaceAll(data.UserName, " ", "_") + "_" + data.UserPhone + "_" + lambdaCategory + "_" + start_minute + "_" + start_hour + "_" + strings.ReplaceAll(week_days, ",", "_") + "_" + strings.ReplaceAll(data.EndTime, ":", "_"))

	ScheduleInput := &scheduler.CreateScheduleInput{
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: "OFF",
		},
		Description:                aws.String(ScheduleDescription),
		Target:                     ScheduleTarget,
		Name:                       ScheduleName,
		ScheduleExpression:         aws.String(ScheduleExpression),
		ScheduleExpressionTimezone: aws.String("America/Lima"),
		State:                      types.ScheduleStateEnabled,
	}

	_, create_err := scheduler_client.CreateSchedule(ctx, ScheduleInput)
	if create_err != nil {
		log.Println("Got an error creating the schedule")
		log.Println(create_err)
		return create_err
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
