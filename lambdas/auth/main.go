package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// No parameters
type MyEvent struct{}

func getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	s3_client := s3.NewFromConfig(cfg)

	return s3_client, err
}

func getWhatsappClient() (*whatsmeow.Client, error) {
	// Environment variables
	db_host := os.Getenv("WHATSAPP_DATABASE_HOST")
	db_password := os.Getenv("WHATSAPP_DATABASE_PASSWORD")

	if db_host == "" || db_password == "" {
		return nil, fmt.Errorf("Environment variables not set")
	}

	// Database connection
	db_port := "5432"
	db_user := "postgres"
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, container_err := sqlstore.New("postgres", "postgres://"+db_user+":"+db_password+"@"+db_host+":"+db_port, dbLog)

	if container_err != nil {
		fmt.Println("Error connecting to database")
		return nil, container_err
	}

	deviceStore, device_err := container.GetFirstDevice()
	if device_err != nil {
		fmt.Println("Error getting first device")
		return nil, device_err
	}

	// WhatsApp client
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	whatsapp_client := whatsmeow.NewClient(deviceStore, clientLog)

	return whatsapp_client, nil
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	// S3 client
	s3_client, s3_err := getS3Client()
	if s3_err != nil {
		fmt.Println("Got an error retrieving the configuration")
		fmt.Println(s3_err)
		return "", s3_err
	}

	// Whatsapp client
	whatsapp_client, whatsapp_err := getWhatsappClient()
	if whatsapp_err != nil {
		fmt.Println("Got an error retrieving the whatsapp client")
		return "", whatsapp_err
	}

	if whatsapp_client.Store.ID == nil {

		// No ID stored, new login
		qrChan, _ := whatsapp_client.GetQRChannel(context.Background())
		connection_err := whatsapp_client.Connect()
		if connection_err != nil {
			fmt.Println("Error connecting to whatsapp client")
			return "", connection_err
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Generate the QR code as png
				var png []byte
				png, qr_err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if qr_err != nil {
					fmt.Println("Error generating QR code")
					return "", qr_err
				}

				// Upload the QR code to S3
				// Assemble the request with the bucket, key, and image body
				input := &s3.PutObjectInput{
					Bucket: aws.String("luna-admin-bucket"),
					Key:    aws.String("WhatsAppQRCode.png"),
					Body:   bytes.NewReader(png),
				}

				// Send the request
				_, s3_err := s3_client.PutObject(context.TODO(), input)
				if s3_err != nil {
					fmt.Println("Error uploading QR code to S3")
					return "", s3_err
				}

			} else {
				fmt.Println("Login event: " + evt.Event)
				return "", fmt.Errorf("Login event: " + evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		conection_err := whatsapp_client.Connect()

		if conection_err != nil {
			fmt.Println("Error connecting to whatsapp client")
			return "", conection_err
		}

		return "Already logged in", nil
	}
	whatsapp_client.Disconnect()
	return "Success", nil
}

func main() {
	lambda.Start(HandleRequest)
}
