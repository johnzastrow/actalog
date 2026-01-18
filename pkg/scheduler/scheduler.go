// Package scheduler provides background job scheduling and session materialization
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/johnzastrow/actalog/pkg/logger"
)

// Job represents a scheduled job that runs at intervals
type Job struct {
	Name     string
	Interval time.Duration
	RunFunc  func(ctx context.Context) error
}

// Scheduler manages background jobs
type Scheduler struct {
	jobs    []*Job
	logger  *logger.Logger
	running bool
	mu      sync.Mutex
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewScheduler creates a new scheduler instance
func NewScheduler(logger *logger.Logger) *Scheduler {
	return &Scheduler{
		jobs:   make([]*Job, 0),
		logger: logger,
	}
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
}

// Start begins running all scheduled jobs
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(ctx, job)
	}

	s.logger.Info("Scheduler started with %d jobs", len(s.jobs))
}

// Stop gracefully stops all running jobs
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("Scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// runJob runs a single job at its configured interval
func (s *Scheduler) runJob(ctx context.Context, job *Job) {
	defer s.wg.Done()

	// Run immediately on start
	if err := job.RunFunc(ctx); err != nil {
		s.logger.Error("Job %s failed: %v", job.Name, err)
	}

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Job %s stopping", job.Name)
			return
		case <-ticker.C:
			s.logger.Debug("Running job: %s", job.Name)
			if err := job.RunFunc(ctx); err != nil {
				s.logger.Error("Job %s failed: %v", job.Name, err)
			}
		}
	}
}

// RunOnce runs a job immediately without scheduling
func (s *Scheduler) RunOnce(ctx context.Context, job *Job) error {
	s.logger.Info("Running job once: %s", job.Name)
	return job.RunFunc(ctx)
}
