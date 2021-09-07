package gpool

import (
	"fmt"
	"testing"
	"time"
)

func run() {
	fmt.Println("running")
	time.Sleep(200 * time.Millisecond)
}

func TestSubmitJobToWorker(t *testing.T) {
	p := NewPool(10, 1000)
	go p.Run()

	index, err := p.Submit(run)
	if err != nil {
		panic(err)
	}
	t.Logf("worker index: %d\n", index)

	for c := 0; c < 10; c++ {
		err = p.SubmitToWorker(run, c)
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(3 * time.Second)

	p.Close()
}

func TestSubmitJob(t *testing.T) {
	p := NewPool(10, 1000)
	go p.Run()

	for c := 0; c < 20; c++ {
		index, err := p.Submit(run)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("submitted to worker: %d", index)
	}

	time.Sleep(3 * time.Second)

	p.Close()
}

func BenchmarkPool(b *testing.B) {
	p := NewPool(100, 10000)
	defer p.Close()

	for n := 0; n < b.N; n++ {
		p.Submit(run)
	}
}
