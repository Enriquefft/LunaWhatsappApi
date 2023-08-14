// package main
//
// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
//
// 	"github.com/aws/aws-lambda-go/lambda"
//
// 	_ "github.com/lib/pq"
// 	"go.mau.fi/whatsmeow"
// 	waProto "go.mau.fi/whatsmeow/binary/proto"
// 	"go.mau.fi/whatsmeow/store/sqlstore"
// 	"go.mau.fi/whatsmeow/types"
// 	waLog "go.mau.fi/whatsmeow/util/log"
//
// 	"google.golang.org/protobuf/proto"
// )
//
// // No parameters
// type MyEvent struct {
// 	Phone   string `json:"phone"`
// 	Message string `json:"message"`
// }
//
// func getWhatsappClient() (*whatsmeow.Client, error) {
// 	// Environment variables
// 	db_host := os.Getenv("WHATSAPP_DATABASE_HOST")
// 	db_password := os.Getenv("WHATSAPP_DATABASE_PASSWORD")
//
// 	if db_host == "" || db_password == "" {
// 		log.Println("WHATSAPP_DATABASE_HOST: " + db_host)
// 		log.Println("WHATSAPP_DATABASE_PASSWORD: " + db_password)
// 		return nil, fmt.Errorf("Environment variables not set")
// 	}
//
// 	// Database connection
// 	db_port := "5432"
// 	db_user := "postgres"
// 	dbLog := waLog.Stdout("Database", "DEBUG", true)
// 	container, container_err := sqlstore.New("postgres", "postgres://"+db_user+":"+db_password+"@"+db_host+":"+db_port, dbLog)
//
// 	if container_err != nil {
// 		log.Println("Error connecting to database")
// 		log.Println(container_err)
// 		return nil, container_err
// 	}
//
// 	deviceStore, device_err := container.GetFirstDevice()
// 	if device_err != nil {
// 		log.Println("Error getting first device")
// 		log.Println(device_err)
// 		return nil, device_err
// 	}
//
// 	// WhatsApp client
// 	clientLog := waLog.Stdout("Client", "DEBUG", true)
// 	whatsapp_client := whatsmeow.NewClient(deviceStore, clientLog)
//
// 	return whatsapp_client, nil
// }
//
// func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
// 	// Whatsapp client
// 	whatsapp_client, whatsapp_err := getWhatsappClient()
// 	if whatsapp_err != nil {
// 		log.Println("Got an error retrieving the whatsapp client")
// 		return "", whatsapp_err
// 	}
//
// 	if whatsapp_client.Store.ID == nil {
// 		log.Println("No client found in the database, please authorize")
// 		return "", fmt.Errorf("No client found in the database, please authorize")
// 	}
//
// 	log.Println("Already logged in, attempting to connect")
// 	conection_err := whatsapp_client.Connect()
//
// 	if conection_err != nil {
// 		log.Println("Error connecting to whatsapp client")
// 		return "", conection_err
// 	}
// 	defer whatsapp_client.Disconnect()
//
// 	err := sendMessage(whatsapp_client, name.Phone, name.Message)
// 	if err != nil {
// 		log.Println("Error sending message")
// 		return "", err
// 	}
//
// 	return "Success", nil
// }
//
// func sendMessage(client *whatsmeow.Client, phone string, message string) error {
// 	log.Println("Sending message to " + phone)
//
// 	// phone must start with +51
// 	if phone[:2] != "+51" {
// 		log.Println("Phone doesn't starts with +51")
// 	}
//
// 	targetJID := types.NewJID(phone, types.DefaultUserServer)
//
// 	response, err := client.SendMessage(context.Background(), targetJID, &waProto.Message{
// 		Conversation: proto.String(message),
// 	})
//
// 	log.Println(response)
//
// 	if err != nil {
// 		log.Println("Error sending message")
// 		return err
// 	}
// 	return nil
// }
//
// func main() {
// 	lambda.Start(HandleRequest)
// }
