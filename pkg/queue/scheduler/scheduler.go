package scheduler

import (
	"WayPointPro/pkg/queue"
	"fmt"
	"time"
)

// Scheduler manages scheduling of tasks
type Scheduler struct {
	tasks []ScheduledTask
}

// ScheduledTask represents a task with an interval
type ScheduledTask struct {
	Interval time.Duration
	Job      queue.Job
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// AddTask adds a task to the scheduler
func (s *Scheduler) AddTask(interval time.Duration, job queue.Job) {
	s.tasks = append(s.tasks, ScheduledTask{
		Interval: interval,
		Job:      job,
	})
}

// Start starts the scheduler and periodically adds tasks to the queue
func (s *Scheduler) Start(q *queue.Queue) {
	for _, task := range s.tasks {
		go func(task ScheduledTask) {
			ticker := time.NewTicker(task.Interval)
			defer ticker.Stop()

			for range ticker.C {
				fmt.Printf("Adding job '%s' to the queue\n", task.Job.Name)
				q.AddJob(task.Job)
			}
		}(task)
	}
}
