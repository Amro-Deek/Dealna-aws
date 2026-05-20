package messaging

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type FCMClient struct {
	app    *firebase.App
	client *messaging.Client
}

func NewFCMClient(ctx context.Context) (*FCMClient, error) {
	// The Firebase SDK automatically uses the GOOGLE_APPLICATION_CREDENTIALS environment variable.
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	return &FCMClient{
		app:    app,
		client: client,
	}, nil
}

// SendSilentPush sends a data-only message that wakes up the app in the background
// without showing a visible notification to the user.
func (f *FCMClient) SendSilentPush(ctx context.Context, token string, data map[string]string) error {
	if token == "" {
		return nil // silently ignore if user has no device token
	}

	// For a purely silent push, we only set Data. We do NOT set Notification.
	message := &messaging.Message{
		Data:  data,
		Token: token,
		// Apns config is needed for iOS background fetch
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
		// Android config for high priority delivery
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}

	_, err := f.client.Send(ctx, message)
	return err
}

// SendVisiblePush sends a message that will show up as a banner notification
func (f *FCMClient) SendVisiblePush(ctx context.Context, token, title, body string, data map[string]string) error {
	if token == "" {
		return nil
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: token,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
	}

	_, err := f.client.Send(ctx, message)
	return err
}
