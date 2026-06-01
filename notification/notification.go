package notification

import (
	"os"

	userClient "github.com/swimresults/user-service/client"
)

var notificationClient *userClient.NotificationClient

func Init() {
	userServiceUrl := os.Getenv("SR_START_USER_URL")
	if userServiceUrl != "" {
		notificationClient = userClient.NewNotificationClient(userServiceUrl)
	}
}
