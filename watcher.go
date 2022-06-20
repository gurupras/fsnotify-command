package fsnotifycmd

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	File string
	Op   fsnotify.Op
}

type Watcher struct {
	watcher  *fsnotify.Watcher
	doneChan chan struct{}
	stopped  bool
}

func (w *Watcher) Stop() {
	if w.stopped {
		return
	}
	defer close(w.doneChan)
	w.stopped = true
	w.doneChan <- struct{}{}
}

func Watch(path string, notifyChan chan<- *Event) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	result := &Watcher{
		watcher:  watcher,
		doneChan: make(chan struct{}),
		stopped:  false,
	}

	go func() {
		defer watcher.Close()
		defer close(notifyChan)
		for {
			log.Infof("Checking if we have events\n")
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Infof("Attempting to write event into notifyChan\n")
				notifyChan <- &Event{
					File: event.Name,
					Op:   event.Op,
				}
			case <-result.doneChan:
				log.Infof("Watcher done\n")
				return
			}
		}
	}()
	if err = watcher.Add(path); err != nil {
		watcher.Close()
		return nil, err
	}
	return result, nil
}
