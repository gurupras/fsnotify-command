package fsnotifycmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWatcherCreateNotification(t *testing.T) {
	require := require.New(t)

	dir, err := ioutil.TempDir("./", "testdir-")
	require.Nil(err)
	defer os.RemoveAll(dir)

	notifyChan := make(chan *Event)
	watcher, err := Watch(dir, notifyChan)
	require.Nil(err)
	require.NotNil(watcher)
	defer watcher.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)
	done := false

	// Set up goroutine to listen for the event
	go func() {
		defer wg.Done()
		for range notifyChan {
			if !done {
				done = true
				break
			}
		}
		log.Infof("Received event")
	}()

	// Create a new file
	f, err := ioutil.TempFile(dir, "file-")
	require.Nil(err)
	defer f.Close()

	wg.Wait()
}

func TestWatcherWriteNotification(t *testing.T) {
	require := require.New(t)

	dir, err := ioutil.TempDir("./", "testdir-")
	require.Nil(err)
	defer os.RemoveAll(dir)

	// Create the file
	f, err := ioutil.TempFile(dir, "file-")
	require.Nil(err)
	defer f.Close()

	notifyChan := make(chan *Event)
	watcher, err := Watch(dir, notifyChan)
	require.Nil(err)
	require.NotNil(watcher)
	defer watcher.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)
	done := false

	// Set up goroutine to listen for the event
	go func() {
		defer wg.Done()
		for range notifyChan {
			if !done {
				done = true
				break
			}
		}
		log.Infof("Received event")
	}()

	// Write to the new file
	_, err = f.Write([]byte("Hello"))
	require.Nil(err)

	wg.Wait()
}

func TestWatcherReplaceNotification(t *testing.T) {
	require := require.New(t)

	dir, err := ioutil.TempDir("./", "testdir-")
	require.Nil(err)
	defer os.RemoveAll(dir)

	// Create the file
	f, err := ioutil.TempFile(dir, "file-")
	require.Nil(err)
	defer f.Close()

	notifyChan := make(chan *Event)
	watcher, err := Watch(dir, notifyChan)
	require.Nil(err)
	require.NotNil(watcher)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Set up goroutine to listen for the event
	go func() {
		defer wg.Done()
		for range notifyChan {
			log.Infof("Received event")
		}
		log.Infof("Done with notify loop")
	}()

	// Create another file in some other location
	f2, err := ioutil.TempFile("", "file-")
	require.Nil(err)
	defer os.Remove(f2.Name())

	_, err = f2.Write([]byte("Hello"))
	require.Nil(err)
	err = f2.Close()
	require.Nil(err)

	log.Infof("About to copy file")
	// Replace the file
	cmd := exec.Command("cp", f2.Name(), f.Name())
	err = cmd.Run()
	require.Nil(err)

	time.Sleep(200 * time.Millisecond)
	watcher.Stop()
	wg.Wait()
	log.Infof("Finished waiting for wg")
}
