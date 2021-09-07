package gpool

import (
	"errors"
	"runtime"
	"sync"
)

var (
	ErrorJobQueueBusy       = errors.New("job queue busy")
	ErrorInvalidWorkerIndex = errors.New("invalid worker index")
)

type Job func()

type Pool struct {
	workerNum  int
	workers    []*worker
	stopChan   chan struct{}
	nextWorker int
	mutex      sync.Mutex
	wg         sync.WaitGroup
}

type worker struct {
	workerIndex     int
	backlogJobQueue chan Job
	pool            *Pool
}

func NewPool(workerNum, backlogJobNum int) (p *Pool) {
	if workerNum <= 0 {
		workerNum = runtime.NumCPU()
	}
	if backlogJobNum <= 0 {
		backlogJobNum = 10000
	}

	p = new(Pool)
	p.workerNum = workerNum
	p.stopChan = make(chan struct{}, 1)
	for counter := 0; counter < workerNum; counter++ {
		p.workers = append(p.workers, p.newWorker(counter, backlogJobNum))
	}

	return
}

func (p *Pool) Run() {
	for _, worker := range p.workers {
		p.wg.Add(1)
		go worker.run(&p.wg)
	}

	p.wg.Wait()
}

func (p *Pool) Close() {
	p.stopChan <- struct{}{}
}

func (p *Pool) Submit(job Job) (workerIndex int, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for counter := 0; counter < p.workerNum; counter++ {
		tmpWorkerIndex := (p.nextWorker + counter) % p.workerNum
		tmpWorker := p.workers[tmpWorkerIndex]
		if len(tmpWorker.backlogJobQueue) == cap(tmpWorker.backlogJobQueue) {
			continue
		}
		tmpWorker.backlogJobQueue <- job
		workerIndex = tmpWorkerIndex
		p.nextWorker = tmpWorkerIndex + 1
		return
	}

	workerIndex = -1
	err = ErrorJobQueueBusy

	return
}

func (p *Pool) SubmitToWorker(job Job, workerIndex int) (err error) {
	if workerIndex < 0 || workerIndex >= p.workerNum {
		err = ErrorInvalidWorkerIndex
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.workers[workerIndex].backlogJobQueue) == cap(p.workers[workerIndex].backlogJobQueue) {
		err = ErrorJobQueueBusy
		return
	}

	p.workers[workerIndex].backlogJobQueue <- job

	return
}

func (p *Pool) newWorker(workerIndex, backlogJobNum int) (w *worker) {
	w = new(worker)
	w.workerIndex = workerIndex
	w.backlogJobQueue = make(chan Job, backlogJobNum)
	w.pool = p

	return
}

func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case jobFunc := <-w.backlogJobQueue:
			jobFunc()
		case <-w.pool.stopChan:
			close(w.backlogJobQueue)
			return
		}
	}
}
