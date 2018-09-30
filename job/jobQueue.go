/*
MIT License

Copyright (c) 2018 Frank Lee

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package job

import (
	"sync/atomic"
	"time"
)

// var queue chan Job

// Job job interface
type Job interface {
	Work()
}

type worker struct {
	workerPool chan chan Job
	jobChan    chan Job
	quit       chan bool
	queue      *JobQueue
}

// Close close this job queue
func (jobQueue *JobQueue) Close() {
	jobQueue.close = true
}

// Submit submit job in queue
func (jobQueue *JobQueue) Submit(job Job) {
	if !jobQueue.close {
		jobQueue.queue <- job
	}
}

// SubmitTimeout submit job in queue with timeout in millisecond
func (jobQueue *JobQueue) SubmitTimeout(job Job, timeout int) bool {
	if timeout <= 0 {
		jobQueue.Submit(job)
		return true
	}
	if !jobQueue.close {
		select {
		case jobQueue.queue <- job:
			return true
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			return false
		}
	}
	return false
}

func newWorker(queue *JobQueue) *worker {
	w := new(worker)
	w.workerPool = queue.workerPool
	w.queue = queue
	w.jobChan = make(chan Job)
	return w
}

func (worker *worker) start() {
	go func() {
	loop:
		for {
			worker.workerPool <- worker.jobChan
			select {
			case job := <-worker.jobChan:
				job.Work()
			case <-worker.quit:
				break loop
			}
		}
	}()
}

func (worker *worker) stop() {
	go func() {
		worker.quit <- true
	}()
}

// JobQueue job queue
type JobQueue struct {
	workerPool chan chan Job
	queue      chan Job
	close      bool
	num        int64
}

// NewJobQueue initialize a JobQueue and start it
// maxWorker max worker counts
// bufferRate  queueSize = maxWorker * bufferRate
func NewJobQueue(maxWorker, bufferRate int) *JobQueue {
	if bufferRate <= 0 {
		bufferRate = 1
	}
	jobQueue := new(JobQueue)
	jobQueue.queue = make(chan Job, maxWorker*bufferRate)
	jobQueue.workerPool = make(chan chan Job, maxWorker)
	jobQueue.close = false
	jobQueue.num = 0
	for i := 0; i < maxWorker; i++ {
		worker := newWorker(jobQueue)
		worker.start()
	}
	go jobQueue.run()
	return jobQueue
}

func (jobQueue *JobQueue) run() {
	for {
		job := <-jobQueue.queue
		atomic.AddInt64(&jobQueue.num, 1)
		jobChan := <-jobQueue.workerPool
		jobChan <- job
	}
}
