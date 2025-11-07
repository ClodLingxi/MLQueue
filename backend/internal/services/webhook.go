package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"MLQueue/internal/config"
	"MLQueue/internal/database"
	"MLQueue/internal/models"
)

type WebhookService struct {
	client *http.Client
}

type WebhookEvent struct {
	Event     string                 `json:"event"`
	TaskID    string                 `json:"task_id"`
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Result    map[string]interface{} `json:"result,omitempty"`
}

// SendWebhook sends webhook notification with retry
func (ws *WebhookService) SendWebhook(event WebhookEvent, userID string) {
	// Get user's webhook configurations
	var webhooks []models.WebhookConfig
	database.DB.Where("user_id = ? AND active = ?", userID, true).Find(&webhooks)

	for _, webhook := range webhooks {
		// Check if webhook is subscribed to this event
		if !ws.isEventSubscribed(webhook.Events, event.Event) {
			continue
		}

		go ws.sendWithRetry(webhook.URL, event, config.AppConfig.Webhook.RetryCount)
	}
}

// sendWithRetry attempts to send webhook with retries
func (ws *WebhookService) sendWithRetry(url string, event WebhookEvent, maxRetries int) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
		if err != nil {
			cancel()
			log.Printf("Failed to create webhook request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "MLQueue-Webhook/1.0")

		resp, err := ws.client.Do(req)
		cancel()

		if err != nil {
			log.Printf("Webhook attempt %d/%d failed for %s: %v", attempt+1, maxRetries+1, url, err)
			continue
		}

		if err := resp.Body.Close(); err != nil {
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook sent successfully to %s", url)
			return
		}

		log.Printf("Webhook attempt %d/%d received status %d for %s", attempt+1, maxRetries+1, resp.StatusCode, url)
	}

	log.Printf("Webhook failed after %d attempts for %s", maxRetries+1, url)
}

// isEventSubscribed checks if webhook is subscribed to event
func (ws *WebhookService) isEventSubscribed(events models.JSONB, eventType string) bool {
	if events == nil {
		return true // Subscribe to all events if not specified
	}

	if eventsList, ok := events["events"].([]interface{}); ok {
		for _, e := range eventsList {
			if e.(string) == eventType {
				return true
			}
		}
		return false
	}

	return true
}

// SendTaskQueued Helper functions to send specific events
func (ws *WebhookService) SendTaskQueued(taskID, userID string) {
	ws.SendWebhook(WebhookEvent{
		Event:     "task.queued",
		TaskID:    taskID,
		Status:    "queued",
		Timestamp: time.Now().Format(time.RFC3339),
	}, userID)
}

func (ws *WebhookService) SendTaskStarted(taskID, userID string) {
	ws.SendWebhook(WebhookEvent{
		Event:     "task.started",
		TaskID:    taskID,
		Status:    "running",
		Timestamp: time.Now().Format(time.RFC3339),
	}, userID)
}

func (ws *WebhookService) SendTaskCompleted(taskID, userID string, result map[string]interface{}) {
	ws.SendWebhook(WebhookEvent{
		Event:     "task.completed",
		TaskID:    taskID,
		Status:    "completed",
		Timestamp: time.Now().Format(time.RFC3339),
		Result:    result,
	}, userID)
}

func (ws *WebhookService) SendTaskFailed(taskID, userID string, errorMsg string) {
	ws.SendWebhook(WebhookEvent{
		Event:     "task.failed",
		TaskID:    taskID,
		Status:    "failed",
		Timestamp: time.Now().Format(time.RFC3339),
		Result:    map[string]interface{}{"error": errorMsg},
	}, userID)
}

func (ws *WebhookService) SendTaskCancelled(taskID, userID string) {
	ws.SendWebhook(WebhookEvent{
		Event:     "task.cancelled",
		TaskID:    taskID,
		Status:    "cancelled",
		Timestamp: time.Now().Format(time.RFC3339),
	}, userID)
}
