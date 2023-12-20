package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"

	"github.com/skip2/go-qrcode"

	// "github.com/aws/aws-lambda-go/lambda"

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

func revokeSession() bool {
	return true
}

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
		log.Println("WHATSAPP_DATABASE_HOST: " + db_host)
		log.Println("WHATSAPP_DATABASE_PASSWORD: " + db_password)
		return nil, errors.New("Environment variables not set")
	}

	// Database connection
	db_port := "5432"
	db_user := "postgres"
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, container_err := sqlstore.New("postgres", "postgres://"+db_user+":"+db_password+"@"+db_host+":"+db_port, dbLog)

	if container_err != nil {
		log.Println("Error connecting to database")
		log.Println(container_err)
		return nil, container_err
	}

	deviceStore, device_err := container.GetFirstDevice()
	if device_err != nil {
		log.Println("Error getting first device")
		log.Println(device_err)
		return nil, device_err
	}

	// WhatsApp client
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	whatsapp_client := whatsmeow.NewClient(deviceStore, clientLog)

	return whatsapp_client, nil
}

func storeHasSession(client *whatsmeow.Client) bool {
	return client.Store.ID != nil
}

func eventHandler(evt interface{}) {
	log.Println("Event handler")
}

func sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	// S3 client
	s3_client, s3_err := getS3Client()
	if s3_err != nil {
		log.Println("Got an error retrieving the configuration")
		log.Println(s3_err)
		return "", s3_err
	}

	// Whatsapp client
	whatsapp_client, whatsapp_err := getWhatsappClient()
	if whatsapp_err != nil {
		log.Println("Got an error retrieving the whatsapp client")
		return "", whatsapp_err
	}
	log.Println("Whatsapp client retrieved")

	if revokeSession() {
		log.Println("Revoking session")
		whatsapp_client.Store.Delete()
	}

	store_has_session := storeHasSession(whatsapp_client)

	if !store_has_session {

		log.Println("No ID stored, new login")

		// No ID stored, new login
		qrChan, _ := whatsapp_client.GetQRChannel(context.Background())
		connection_err := whatsapp_client.Connect()
		if connection_err != nil {
			log.Println("Error connecting to whatsapp client")
			return "", connection_err
		}
		// defer fmt.Println("Disconnecting from wsp client")
		// defer whatsapp_client.Disconnect()

		for evt := range qrChan {
			if evt.Event == "code" {

				// Generating new code
				log.Println("Generating new code")

				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)

				// Generate the QR code as png
				var png []byte
				png, qr_err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if qr_err != nil {
					log.Println("Error generating QR code")
					return "", qr_err
				}

				// Replace the QR code in S3
				input := &s3.PutObjectInput{
					Bucket: aws.String("luna-auth-bucket"),
					Key:    aws.String("WhatsAppQRCode.png"),
					Body:   bytes.NewReader(png),
				}

				// Send the request
				_, s3_err := s3_client.PutObject(context.TODO(), input)
				if s3_err != nil {
					log.Println("Error uploading QR code to S3")
					return "", s3_err
				} else {
					log.Println("QR code uploaded to S3")
				}

			} else {
				log.Println("Login event: " + evt.Event)
			}
		}
	} else {
		log.Println("Already logged in, attempting to connect")
		conection_err := whatsapp_client.Connect()

		if conection_err != nil {
			log.Println("Error connecting to whatsapp client")
			return "", conection_err
		}
		defer whatsapp_client.Disconnect()
		return "Already logged in", nil
	}

	sleep(5)

	// whatsapp_client.Disconnect()
	return "Success", nil
}

func main() {
	res, err := HandleRequest(context.TODO(), MyEvent{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println(res)
	}
}
