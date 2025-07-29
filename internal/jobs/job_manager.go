package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	JobPriorityLow    JobPriority = 0
	JobPriorityNormal JobPriority = 1
	JobPriorityHigh   JobPriority = 2
	JobPriorityUrgent JobPriority = 3
)

// Job represents an async job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    JobPriority            `json:"priority"`
	Status      JobStatus              `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// JobManager manages async job processing
type JobManager struct {
	handlers      map[string]JobHandler
	jobs          map[string]*Job
	queue         chan *Job
	workers       int
	maxQueueSize  int
	mu            sync.RWMutex
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *log.Logger
}

// JobManagerConfig holds configuration for job manager
type JobManagerConfig struct {
	Workers      int
	MaxQueueSize int
	Logger       *log.Logger
}

// NewJobManager creates a new job manager
func NewJobManager(config JobManagerConfig) *JobManager {
	if config.Workers <= 0 {
		config.Workers = 4
	}
	if config.MaxQueueSize <= 0 {
		config.MaxQueueSize = 1000
	}
	if config.Logger == nil {
		config.Logger = log.Default()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	jm := &JobManager{
		handlers:     make(map[string]JobHandler),
		jobs:         make(map[string]*Job),
		queue:        make(chan *Job, config.MaxQueueSize),
		workers:      config.Workers,
		maxQueueSize: config.MaxQueueSize,
		ctx:          ctx,
		cancel:       cancel,
		logger:       config.Logger,
	}
	
	// Start workers
	jm.startWorkers()
	
	return jm
}

// RegisterHandler registers a job handler
func (jm *JobManager) RegisterHandler(jobType string, handler JobHandler) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.handlers[jobType] = handler
}

// SubmitJob submits a job for processing
func (jm *JobManager) SubmitJob(job *Job) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}
	if job.Status == "" {
		job.Status = JobStatusPending
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	
	jm.mu.Lock()
	jm.jobs[job.ID] = job
	jm.mu.Unlock()
	
	select {
	case jm.queue <- job:
		jm.logger.Printf("Job %s submitted successfully", job.ID)
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("job queue is full")
	}
}

// GetJob returns a job by ID
func (jm *JobManager) GetJob(id string) (*Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	job, exists := jm.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	
	return job, nil
}

// GetJobsByStatus returns jobs with a specific status
func (jm *JobManager) GetJobsByStatus(status JobStatus) []*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	var jobs []*Job
	for _, job := range jm.jobs {
		if job.Status == status {
			jobs = append(jobs, job)
		}
	}
	
	return jobs
}

// CancelJob cancels a pending or running job
func (jm *JobManager) CancelJob(id string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	
	job, exists := jm.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}
	
	if job.Status != JobStatusPending && job.Status != JobStatusRunning {
		return fmt.Errorf("job %s cannot be cancelled (status: %s)", id, job.Status)
	}
	
	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now
	
	return nil
}

// startWorkers starts the worker goroutines
func (jm *JobManager) startWorkers() {
	for i := 0; i < jm.workers; i++ {
		jm.wg.Add(1)
		go jm.worker(i)
	}
}

// worker processes jobs from the queue
func (jm *JobManager) worker(id int) {
	defer jm.wg.Done()
	
	jm.logger.Printf("Worker %d started", id)
	
	for {
		select {
		case <-jm.ctx.Done():
			jm.logger.Printf("Worker %d shutting down", id)
			return
			
		case job := <-jm.queue:
			jm.processJob(job)
		}
	}
}

// processJob processes a single job
func (jm *JobManager) processJob(job *Job) {
	jm.logger.Printf("Processing job %s (type: %s)", job.ID, job.Type)
	
	// Update job status
	jm.updateJobStatus(job, JobStatusRunning)
	now := time.Now()
	job.StartedAt = &now
	
	// Get handler
	jm.mu.RLock()
	handler, exists := jm.handlers[job.Type]
	jm.mu.RUnlock()
	
	if !exists {
		jm.handleJobError(job, fmt.Errorf("no handler registered for job type: %s", job.Type))
		return
	}
	
	// Create job context with timeout
	ctx, cancel := context.WithTimeout(jm.ctx, 30*time.Minute)
	defer cancel()
	
	// Execute job
	err := handler(ctx, job)
	
	if err != nil {
		jm.handleJobError(job, err)
	} else {
		jm.handleJobSuccess(job)
	}
}

// updateJobStatus updates the status of a job
func (jm *JobManager) updateJobStatus(job *Job, status JobStatus) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	job.Status = status
}

// handleJobError handles job execution error
func (jm *JobManager) handleJobError(job *Job, err error) {
	job.Error = err.Error()
	job.RetryCount++
	
	if job.RetryCount < job.MaxRetries {
		// Retry the job
		jm.logger.Printf("Job %s failed, retrying (%d/%d): %v", job.ID, job.RetryCount, job.MaxRetries, err)
		job.Status = JobStatusPending
		
		// Add back to queue with exponential backoff
		go func() {
			delay := time.Duration(job.RetryCount) * time.Minute
			time.Sleep(delay)
			jm.queue <- job
		}()
	} else {
		// Max retries exceeded
		jm.logger.Printf("Job %s failed after %d retries: %v", job.ID, job.MaxRetries, err)
		job.Status = JobStatusFailed
		now := time.Now()
		job.CompletedAt = &now
	}
	
	jm.mu.Lock()
	jm.jobs[job.ID] = job
	jm.mu.Unlock()
}

// handleJobSuccess handles successful job completion
func (jm *JobManager) handleJobSuccess(job *Job) {
	jm.logger.Printf("Job %s completed successfully", job.ID)
	
	job.Status = JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	
	jm.mu.Lock()
	jm.jobs[job.ID] = job
	jm.mu.Unlock()
}

// Shutdown gracefully shuts down the job manager
func (jm *JobManager) Shutdown() {
	jm.logger.Println("Shutting down job manager...")
	
	// Cancel context to signal workers to stop
	jm.cancel()
	
	// Wait for all workers to finish
	jm.wg.Wait()
	
	// Close the queue
	close(jm.queue)
	
	jm.logger.Println("Job manager shut down complete")
}

// GetStats returns job manager statistics
func (jm *JobManager) GetStats() JobManagerStats {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	stats := JobManagerStats{
		TotalJobs:    len(jm.jobs),
		QueueSize:    len(jm.queue),
		Workers:      jm.workers,
		MaxQueueSize: jm.maxQueueSize,
		JobsByStatus: make(map[JobStatus]int),
	}
	
	for _, job := range jm.jobs {
		stats.JobsByStatus[job.Status]++
	}
	
	return stats
}

// JobManagerStats holds job manager statistics
type JobManagerStats struct {
	TotalJobs    int                  `json:"total_jobs"`
	QueueSize    int                  `json:"queue_size"`
	Workers      int                  `json:"workers"`
	MaxQueueSize int                  `json:"max_queue_size"`
	JobsByStatus map[JobStatus]int    `json:"jobs_by_status"`
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// JobQueue interface for different queue implementations
type JobQueue interface {
	Push(job *Job) error
	Pop() (*Job, error)
	Size() int
	Close() error
}

// PriorityQueue implements a priority-based job queue
type PriorityQueue struct {
	queues map[JobPriority]chan *Job
	mu     sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(size int) *PriorityQueue {
	pq := &PriorityQueue{
		queues: make(map[JobPriority]chan *Job),
	}
	
	// Initialize queues for each priority level
	for _, priority := range []JobPriority{JobPriorityUrgent, JobPriorityHigh, JobPriorityNormal, JobPriorityLow} {
		pq.queues[priority] = make(chan *Job, size/4)
	}
	
	return pq
}

// Push adds a job to the appropriate priority queue
func (pq *PriorityQueue) Push(job *Job) error {
	queue, exists := pq.queues[job.Priority]
	if !exists {
		return fmt.Errorf("invalid job priority: %d", job.Priority)
	}
	
	select {
	case queue <- job:
		return nil
	default:
		return fmt.Errorf("priority queue %d is full", job.Priority)
	}
}

// Pop retrieves the highest priority job
func (pq *PriorityQueue) Pop() (*Job, error) {
	// Check queues in priority order
	for _, priority := range []JobPriority{JobPriorityUrgent, JobPriorityHigh, JobPriorityNormal, JobPriorityLow} {
		select {
		case job := <-pq.queues[priority]:
			return job, nil
		default:
			continue
		}
	}
	
	return nil, fmt.Errorf("all queues are empty")
}

// Size returns the total number of jobs in all queues
func (pq *PriorityQueue) Size() int {
	total := 0
	for _, queue := range pq.queues {
		total += len(queue)
	}
	return total
}

// Close closes all priority queues
func (pq *PriorityQueue) Close() error {
	for _, queue := range pq.queues {
		close(queue)
	}
	return nil
}