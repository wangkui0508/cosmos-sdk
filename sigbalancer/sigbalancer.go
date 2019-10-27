package sigbalancer

import (
	"sync/atomic"

	"github.com/tendermint/tendermint/crypto"
)

var requestChan = make(chan Request, nRequester)
var SyncChan = make(chan int64, 100000)

var TotalRequest int64 = 0
var TotalFinished int64 = 0

var ErrorCount int64 = 0

func SendRequest(pubkey crypto.PubKey, msg []byte, sig []byte) {
	req := Request{fn: func() {
			ok := pubkey.VerifyBytes(msg, sig)
			newFinished := atomic.AddInt64(&TotalFinished, 1)
			if !ok {
				atomic.AddInt64(&ErrorCount, 1)
			}
			//fmt.Printf("Finished Verify %v pending %d ErrorCount %d\n", ok, newFinished, newCount)
			if newFinished == TotalRequest {
				SyncChan <- newFinished
			}
		},
	}
	atomic.AddInt64(&TotalRequest, 1)
	requestChan <- req
}

var balancer = NewBalancer()

func init() {
	go balancer.balance(requestChan)
}

