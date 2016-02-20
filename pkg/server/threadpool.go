/*
Copyright © 2014–6 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server

import (
	log "github.com/Sirupsen/logrus"
	"runtime"
)

// Create a global pool.
var globalPool = StartPool(runtime.NumCPU())

// Submit submits a job to the default (global) pool.
func Submit(job func()) {
	globalPool.Submit(job)
}

// ThreadPool distributes jobs to workers.
type ThreadPool interface {
	// Submit takes the job passed as input and schedules it for execution on an
	// available goroutine.
	Submit(func())

	// Quit terminates the workers and then the threadpool.
	Quit()
}

type threadPool struct {
	workQueue   chan func()
	workerQueue chan chan func()
	quit        chan bool
}

// StartPool starts a new thread pool.
func StartPool(numWorkers int) ThreadPool {
	log.Printf("Starting a thread pool with %v workers", numWorkers)
	workerQueue := make(chan chan func(), numWorkers)
	p := threadPool{
		workQueue:   make(chan func(), 100),
		workerQueue: workerQueue,
		quit:        make(chan bool),
	}
	for i := 0; i < numWorkers; i++ {
		worker := newWorker(i+1, workerQueue)
		worker.Start()
	}

	// And now the main loop
	go func() {
		for {
			select {
			case job := <-p.workQueue:
				// Received a job; dispatch it.
				go func() {
					worker := <-p.workerQueue
					worker <- job
				}()
			case <-p.quit:
				// DOME: kill all the workers.
				return
			}
		}
	}()

	return &p
}

func (p *threadPool) Submit(job func()) {
	p.workQueue <- job
}

func (p *threadPool) Quit() {
	go func() {
		p.quit <- true
	}()
}

// Worker is our worker.
type Worker struct {
	ID          int
	Work        chan func()
	WorkerQueue chan chan func()
	Quit        chan bool
}

// newWorker creates and returns a new worker.
func newWorker(id int, workerQueue chan chan func()) Worker {
	return Worker{
		ID:          id,
		Work:        make(chan func()),
		WorkerQueue: workerQueue,
		Quit:        make(chan bool),
	}
}

// Start starts the specified worker.
func (w *Worker) Start() {
	go func() {
		for {
			// Add this worker to the list of available workers.
			w.WorkerQueue <- w.Work
			select {
			// We have work to do.
			case job := <-w.Work:
				job()
			// Quit.
			case <-w.Quit:
				return
			}
		}
	}()
}

// Stop signals the specified worker to stop once it has finished the current
// job.
func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
