### how to install

---

`go get -u -v github.com/ClaytonXia/gpool`

### how to use

---

```go
	// init the pool with total goroutine num and backlog job num of single goroutine
	p := gpool.NewPool(10, 1000)
	go p.Run()
	defer p.Close()

	func jobFunc() {
		// job logic
	}

	// submit to worker and return worker index
	workerIndex, err := p.Submit(jobFunc)
	if err == gpool.ErrorJobQueueBusy {
		// process busy job error
	}

	// you can also submit job to certain worker by passing the worker index param
	p.SubmitToWorker(jobFunc, workerIndex)
```
