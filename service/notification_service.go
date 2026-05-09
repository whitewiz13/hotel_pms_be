package service

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/hotelpms/backend/repository"
	"google.golang.org/api/option"
)

const notificationWorkers = 50

type notificationJob struct {
	tokens []string
	title  string
	body   string
	data   map[string]string
}

// NotificationService sends push notifications via Firebase Cloud Messaging.
type NotificationService struct {
	client       *messaging.Client
	fcmTokenRepo *repository.FCMTokenRepository
	jobCh        chan notificationJob
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

	svc := &NotificationService{
		fcmTokenRepo: fcmTokenRepo,
		jobCh:        make(chan notificationJob, 500),
	}

	if err != nil {
		log.Printf("[FCM] Firebase init failed: %v — notifications disabled", err)
		svc.startWorkers()
		return svc
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("[FCM] Messaging client failed: %v — notifications disabled", err)
		svc.startWorkers()
		return svc
	}

	svc.client = client
	svc.startWorkers()
	log.Println("[FCM] Firebase Cloud Messaging initialised")
	return svc
}

// startWorkers launches a fixed pool of goroutines to process notification jobs.
func (s *NotificationService) startWorkers() {
	for i := 0; i < notificationWorkers; i++ {
		go func() {
			for job := range s.jobCh {
				s.doSend(job.tokens, job.title, job.body, job.data)
			}
		}()
	}
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

	s.enqueue(deviceTokens, title, body, data)
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

	s.enqueue(tokens, title, body, data)
}

func (s *NotificationService) enqueue(tokens []string, title, body string, data map[string]string) {
	select {
	case s.jobCh <- notificationJob{tokens: tokens, title: title, body: body, data: data}:
	default:
		log.Printf("[FCM] notification queue full, dropping notification: %s", title)
	}
}

func (s *NotificationService) doSend(tokens []string, title, body string, data map[string]string) {
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
