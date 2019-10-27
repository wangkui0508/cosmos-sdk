package sigbalancer

import (
	"math/rand"
	"time"
	"testing"
)


// Simulation of some work: just sleep for a while and report how long.
func op() int {
	n := rand.Int63n(int64(time.Second))
	time.Sleep(time.Duration(nWorker * n))
	return int(n)
}
func requester(requestChan chan Request) {
	for {
		time.Sleep(time.Duration(rand.Int63n(int64(nWorker * 2 * time.Second))))
		requestChan <- Request{op}
	}
}

func Test1(t *testing.T) {
	requestChan := make(chan Request)
	for i := 0; i < nRequester; i++ {
		go requester(requestChan)
	}
	NewBalancer().balance(requestChan)
}
