package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/google/shlex"
	"github.com/gurupras/fsnotifycmd"
	log "github.com/sirupsen/logrus"
)

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if strings.Compare(val, "") == 0 {
		val = defaultValue
	}
	return val
}

type WatchData struct {
	Path    string `json:"path"`
	Command string `json:"command"`
}

var (
	defaultWatch   = getEnv("FSNOTIFY_CMD_WATCH", "[]")
	defaultVerbose = getEnv("FSNOTIFY_CMD_VERBOSE", "false")

	watch   = kingpin.Flag("watch", "Watch data").Short('w').Default(defaultWatch).String()
	verbose = kingpin.Flag("verbose", "Verbose logs").Short('v').Default(defaultVerbose).Bool()
)

func main() {
	kingpin.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	watchObj := make([]*WatchData, 0)
	err := json.Unmarshal([]byte(*watch), &watchObj)
	if err != nil {
		log.Fatalf("Failed to parse --watch flag: %v", err)
	}

	wg := sync.WaitGroup{}
	for _, entry := range watchObj {
		wg.Add(1)
		go func(entry *WatchData) {
			defer wg.Done()
			tokens, err := shlex.Split(entry.Command)
			if err != nil {
				log.Fatalf("Failed to split command: %v", err)
			}
			command := tokens[0]
			args := tokens[1:]

			notifyChan := make(chan *fsnotifycmd.Event)
			w, err := fsnotifycmd.Watch(entry.Path, notifyChan)
			if err != nil {
				log.Fatalf("Failed to watch '%v': %v", entry.Path, err)
			}
			defer w.Stop()
			// We need to convert our notifyChan to a chan interface{} before we can debounce it
			iChan := make(chan interface{})
			go func() {
				defer close(iChan)
				for o := range notifyChan {
					iChan <- o
				}
			}()
			dChan := fsnotifycmd.DebounceChan(1*time.Second, iChan)
			for range dChan {
				cmd := exec.Command(command, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				log.Debugf("Running command '%v'\n", entry.Command)
				err = cmd.Run()
				if err != nil {
					log.Errorf("Failed to run command '%v': %v\n", entry.Command, err)
				}
			}
		}(entry)
	}

	wg.Wait()
}
