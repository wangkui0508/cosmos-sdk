package sigbalancer

import (
	"container/heap"
	"fmt"
)

const nWorker = 10
const nRequester = 100

type Request struct {
	fn func()
}

type Worker struct {
	id        int
	requests chan Request
	pending  int
}

func (w *Worker) work(done chan *Worker) {
	for {
		req := <-w.requests
		//fmt.Printf("Get a request at %d len: %d\n", w.id, len(w.requests))
		req.fn()
		done <- w
	}
}

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].id = i
	a[j].id = j
}

func (p *Pool) Push(x interface{}) {
	a := *p
	n := len(a)
	a = a[0 : n+1]
	w := x.(*Worker)
	a[n] = w
	w.id = n
	*p = a
}

func (p *Pool) Pop() interface{} {
	a := *p
	*p = a[0 : len(a)-1]
	w := a[len(a)-1]
	w.id = -1 // for safety
	return w
}

type Balancer struct {
	pool Pool
	done chan *Worker
}

func NewBalancer() *Balancer {
	done := make(chan *Worker, nWorker)
	b := &Balancer{make(Pool, 0, nWorker), done}
	for i := 0; i < nWorker; i++ {
		w := &Worker{requests: make(chan Request, 8)}
		heap.Push(&b.pool, w)
		go w.work(b.done)
	}
	return b
}

func (b *Balancer) balance(requestChan chan Request) {
	for {
		select {
		case req := <-requestChan:
			b.dispatch(req)
		case w := <-b.done:
			b.completed(w)
		}
		//b.printStatus()
	}
}

func (b *Balancer) printStatus() {
	sum := 0
	sumsq := 0
	for _, w := range b.pool {
		fmt.Printf("%d ", w.pending)
		sum += w.pending
		sumsq += w.pending * w.pending
	}
	avg := float64(sum) / float64(len(b.pool))
	variance := float64(sumsq)/float64(len(b.pool)) - avg*avg
	fmt.Printf(" %.2f %.2f\n", avg, variance)
}

func (b *Balancer) dispatch(req Request) {
	// Grab least loaded worker
	w := heap.Pop(&b.pool).(*Worker)
	w.requests <- req
	w.pending++
	// Put it back into heap while it is working
	heap.Push(&b.pool, w)
}

func (b *Balancer) completed(w *Worker) {
	w.pending--
	// remove from heap
	heap.Remove(&b.pool, w.id)
	// Put it back
	heap.Push(&b.pool, w)
}

