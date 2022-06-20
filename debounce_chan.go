package fsnotifycmd

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func DebounceChan(interval time.Duration, inputChan <-chan interface{}) chan interface{} {
	ret := make(chan interface{})
	var timerFunc *time.Timer

	var item interface{}

	done := false
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	timerChan := make(chan struct{})

	go func() {
		defer wg.Done()
		for range timerChan {
		}
	}()

	go func() {
		defer func() {
			mutex.Lock()
			if timerFunc != nil {
				done = true
				mutex.Unlock()
				wg.Wait()
			} else {
				mutex.Unlock()
			}
			close(ret)
		}()
		for obj := range inputChan {
			log.Debugf("Got message from inputChan")
			mutex.Lock()
			item = obj
			if timerFunc != nil {
				timerFunc.Stop()
				log.Debugf("Reset timer")
				timerFunc = nil
			}
			timerFunc = time.AfterFunc(interval, func() {
				mutex.Lock()
				log.Debugf("Writing message to debouncedChan")
				ret <- item
				log.Debugf("Wrote message to debouncedChan")
				timerChan <- struct{}{}
				defer mutex.Unlock()
				if done {
					close(timerChan)
				}
				timerFunc = nil
			})
			mutex.Unlock()
		}
	}()
	return ret
}
