package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"MLQueue/internal/database"
	"MLQueue/internal/models"

	"github.com/redis/go-redis/v9"
)

const (
	TaskQueueKey    = "mlqueue:tasks"
	TaskQueueSetKey = "mlqueue:tasks:set"
)

type Manager struct {
	redis       *redis.Client
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	paused      bool
	mu          sync.RWMutex
}

func NewQueueManager(workerCount int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		redis:       database.RedisClient,
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		paused:      false,
	}
}

// Start begins processing queue with multiple workers
func (qm *Manager) Start() {
	log.Printf("Starting queue manager with %d workers", qm.workerCount)

	for i := 0; i < qm.workerCount; i++ {
		qm.wg.Add(1)
		go qm.worker(i)
	}
}

// worker processes tasks from queue
func (qm *Manager) worker(id int) {
	defer qm.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-qm.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		default:
			qm.mu.RLock()
			isPaused := qm.paused
			qm.mu.RUnlock()

			if isPaused {
				time.Sleep(1 * time.Second)
				continue
			}

			// Use BZPOPMIN for blocking pop with timeout
			result, err := qm.redis.BZPopMin(qm.ctx, 2*time.Second, TaskQueueKey).Result()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				log.Printf("Worker %d: error popping from queue: %v", id, err)
				continue
			}

			taskID := result.Member.(string)
			qm.processTask(id, taskID)
		}
	}
}

// processTask handles individual task execution
func (qm *Manager) processTask(workerID int, taskID string) {
	log.Printf("Worker %d: processing task %s", workerID, taskID)

	// Get task from database
	var task models.Task
	if err := database.DB.First(&task, "id = ?", taskID).Error; err != nil {
		log.Printf("Worker %d: failed to load task %s: %v", workerID, taskID, err)
		return
	}

	// Update status to running
	now := time.Now()
	task.Status = models.TaskStatusRunning
	task.StartedAt = &now

	if err := database.DB.Save(&task).Error; err != nil {
		log.Printf("Worker %d: failed to update task status: %v", workerID, err)
		return
	}

	// Notify status change
	qm.publishStatusChange(taskID, string(models.TaskStatusRunning))

	// Simulate task processing (in real scenario, this would execute the actual training)
	// For demonstration, we'll just wait and mark as completed
	time.Sleep(time.Duration(5+workerID) * time.Second)

	// Mark as completed
	completedAt := time.Now()
	task.Status = models.TaskStatusCompleted
	task.CompletedAt = &completedAt
	task.Result = models.JSONB{
		"completed_by_worker": workerID,
		"duration_seconds":    completedAt.Sub(*task.StartedAt).Seconds(),
	}

	if err := database.DB.Save(&task).Error; err != nil {
		log.Printf("Worker %d: failed to complete task: %v", workerID, err)
		return
	}

	// Remove from set
	qm.redis.SRem(qm.ctx, TaskQueueSetKey, taskID)

	// Notify completion
	qm.publishStatusChange(taskID, string(models.TaskStatusCompleted))

	log.Printf("Worker %d: completed task %s", workerID, taskID)
}

// EnqueueTask adds a task to the queue
func (qm *Manager) EnqueueTask(taskID string, priority float64) error {
	// Add to sorted set (priority queue)
	if err := qm.redis.ZAdd(qm.ctx, TaskQueueKey, redis.Z{
		Score:  -priority, // Negative for descending order
		Member: taskID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Add to set for tracking
	if err := qm.redis.SAdd(qm.ctx, TaskQueueSetKey, taskID).Err(); err != nil {
		return fmt.Errorf("failed to add task to set: %w", err)
	}

	return nil
}

// GetQueueLength returns current queue size
func (qm *Manager) GetQueueLength() (int64, error) {
	return qm.redis.ZCard(qm.ctx, TaskQueueKey).Result()
}

// GetQueuePosition returns task position in queue
func (qm *Manager) GetQueuePosition(taskID string) (int64, error) {
	rank, err := qm.redis.ZRank(qm.ctx, TaskQueueKey, taskID).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return rank + 1, nil
}

// UpdatePriority changes task priority in queue
func (qm *Manager) UpdatePriority(taskID string, newPriority float64) error {
	return qm.redis.ZAdd(qm.ctx, TaskQueueKey, redis.Z{
		Score:  -newPriority,
		Member: taskID,
	}).Err()
}

// RemoveTask removes a task from queue
func (qm *Manager) RemoveTask(taskID string) error {
	if err := qm.redis.ZRem(qm.ctx, TaskQueueKey, taskID).Err(); err != nil {
		return err
	}
	return qm.redis.SRem(qm.ctx, TaskQueueSetKey, taskID).Err()
}

// Pause pauses queue processing
func (qm *Manager) Pause() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.paused = true
	log.Println("Queue paused")
}

// Resume resumes queue processing
func (qm *Manager) Resume() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.paused = false
	log.Println("Queue resumed")
}

// IsPaused returns current pause status
func (qm *Manager) IsPaused() bool {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.paused
}

// publishStatusChange publishes task status changes to Redis pub/sub
func (qm *Manager) publishStatusChange(taskID, status string) {
	message := map[string]string{
		"task_id": taskID,
		"status":  status,
		"time":    time.Now().Format(time.RFC3339),
	}

	data, _ := json.Marshal(message)
	qm.redis.Publish(qm.ctx, "task:status:"+taskID, data)
	qm.redis.Publish(qm.ctx, "task:status:all", data)
}

// Stop gracefully stops the queue manager
func (qm *Manager) Stop() {
	log.Println("Stopping queue manager...")
	qm.cancel()
	qm.wg.Wait()
	log.Println("Queue manager stopped")
}
