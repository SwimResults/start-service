package notification

import (
	"fmt"
	userClient "github.com/swimresults/user-service/client"
	"os"
)

var notificationClient *userClient.NotificationClient
var serviceKey string

func Init() {

	serviceKey = os.Getenv("SR_SERVICE_KEY")

	if serviceKey == "" {
		fmt.Println("no security for inter-service communication given! Please set SR_SERVICE_KEY.")
	}

	userServiceUrl := os.Getenv("SR_START_USER_URL")
	if userServiceUrl != "" {
		notificationClient = userClient.NewNotificationClient(userServiceUrl)
	}
}
