package queue

import (
	"fmt"
	"sync"
)

// Job represents a task in the queue
type Job struct {
	ID      int
	Name    string
	Execute func() // Function to execute the job
}

// Queue represents the job queue
type Queue struct {
	JobQueue chan Job // Channel to hold jobs
	Workers  int      // Number of workers
	wg       sync.WaitGroup
}

// NewQueue creates a new job queue
func NewQueue(workers int) *Queue {
	return &Queue{
		JobQueue: make(chan Job, 100), // Buffer size of 100 jobs
		Workers:  workers,
	}
}

// Start initializes the workers to process jobs from the queue
func (q *Queue) Start() {
	for i := 0; i < q.Workers; i++ {
		go func(workerID int) {
			for job := range q.JobQueue {
				fmt.Printf("Worker %d processing job: %s\n", workerID, job.Name)
				job.Execute()
				q.wg.Done()
			}
		}(i)
	}
}

// AddJob adds a job to the queue
func (q *Queue) AddJob(job Job) {
	q.wg.Add(1)
	q.JobQueue <- job
}

// Wait waits for all jobs to be processed
func (q *Queue) Wait() {
	q.wg.Wait()
	close(q.JobQueue) // Close the job queue after processing all jobs
}
