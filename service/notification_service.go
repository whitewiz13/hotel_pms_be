package service

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/hotelpms/backend/repository"
	"google.golang.org/api/option"
)

// NotificationService sends push notifications via Firebase Cloud Messaging.
type NotificationService struct {
	client       *messaging.Client
	fcmTokenRepo *repository.FCMTokenRepository
}

// NewNotificationService initialises Firebase and returns the service.
// It prefers credentialsJSON (raw JSON string, for production/Render),
// then credentialsFile (local file path), then GOOGLE_APPLICATION_CREDENTIALS env var.
func NewNotificationService(credentialsFile, credentialsJSON string, fcmTokenRepo *repository.FCMTokenRepository) *NotificationService {
	ctx := context.Background()

	var app *firebase.App
	var err error

	if credentialsJSON != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credentialsJSON)))
	} else if credentialsFile != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		log.Printf("[FCM] Firebase init failed: %v — notifications disabled", err)
		return &NotificationService{fcmTokenRepo: fcmTokenRepo}
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("[FCM] Messaging client failed: %v — notifications disabled", err)
		return &NotificationService{fcmTokenRepo: fcmTokenRepo}
	}

	log.Println("[FCM] Firebase Cloud Messaging initialised")
	return &NotificationService{client: client, fcmTokenRepo: fcmTokenRepo}
}

// SendToUser sends a notification to all devices of a specific user.
func (s *NotificationService) SendToUser(userID, title, body string, data map[string]string) {
	if s.client == nil {
		return
	}

	tokens, err := s.fcmTokenRepo.FindByUserID(userID)
	if err != nil || len(tokens) == 0 {
		return
	}

	deviceTokens := make([]string, len(tokens))
	for i, t := range tokens {
		deviceTokens[i] = t.DeviceToken
	}

	s.sendToTokens(deviceTokens, title, body, data)
}

// SendToHotelStaff sends a notification to all hotel staff who have at least
// one of the given permission codes. Hotel admins are always included.
func (s *NotificationService) SendToHotelStaff(hotelID, title, body string, data map[string]string, permissions ...string) {
	if s.client == nil {
		return
	}

	tokens, err := s.fcmTokenRepo.FindTokensByHotelAndPermission(hotelID, permissions...)
	if err != nil || len(tokens) == 0 {
		return
	}

	s.sendToTokens(tokens, title, body, data)
}

func (s *NotificationService) sendToTokens(tokens []string, title, body string, data map[string]string) {
	if len(tokens) == 0 {
		return
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	resp, err := s.client.SendEachForMulticast(context.Background(), msg)
	if err != nil {
		log.Printf("[FCM] SendEachForMulticast error: %v", err)
		return
	}

	if resp.FailureCount > 0 {
		log.Printf("[FCM] %d/%d messages failed", resp.FailureCount, len(tokens))
	}
}
