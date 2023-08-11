package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)

	// Read environment variables
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}

	db_host := os.Getenv("WHATSAPP_DATABASE_HOST")
	// "luna-whatsapp-db.cnnofz8s8xzo.us-east-1.rds.amazonaws.com"
	db_password := os.Getenv("WHATSAPP_DATABASE_PASSWORD")
	// "YSBk04pnqBiqdgTxCLRa"

	db_port := "5432"
	db_user := "postgres"

	// db_host and db_password must be set as environment variables
	if db_host == "" || db_password == "" {
		fmt.Println("WHATSAPP_DATABASE_HOST:\t" + db_host)
		fmt.Println("WHATSAPP_DATABASE_PASSWORD:\t" + db_password)
		panic("Missing database environment variables")
	}

	container, err := sqlstore.New("postgres", "postgres://"+db_user+":"+db_password+"@"+db_host+":"+db_port, dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				// fmt.Println("QR code:", evt.Code)
				qrterminal.Generate(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
