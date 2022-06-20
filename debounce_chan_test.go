package fsnotifycmd

import (
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestDebounceChanDebouncesMessages(t *testing.T) {
	require := require.New(t)

	interval := 10 * time.Millisecond
	numMessages := 10
	expectedCount := 1

	input := make(chan interface{})
	debouncedChan := DebounceChan(interval, input)

	count := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range debouncedChan {
			log.Debugf("Received debounced event")
			count++
		}
	}()

	for idx := 0; idx < numMessages; idx++ {
		log.Debugf("Wrote message to inputChan")
		input <- struct{}{}
	}
	close(input)

	wg.Wait()
	require.Equal(expectedCount, count)
}

func TestDebounceChanDebouncesMessagesBasedOnInterval(t *testing.T) {
	require := require.New(t)

	interval := 10 * time.Millisecond
	numMessages := 10
	expectedCount := 10

	input := make(chan interface{})
	debouncedChan := DebounceChan(interval, input)

	count := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range debouncedChan {
			log.Debugf("Received debounced event")
			count++
		}
	}()

	for idx := 0; idx < numMessages; idx++ {
		log.Debugf("Wrote message to inputChan")
		input <- struct{}{}
		time.Sleep(interval + 1*time.Millisecond)
	}
	close(input)

	wg.Wait()
	require.Equal(expectedCount, count)
}
