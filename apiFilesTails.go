package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

// FilesTails ...
type FilesTails struct {
	context   context.Context
	eventLoop jsEngine.EventLoop
	runtime   *goja.Runtime
}

// FilesTailsWatcherConfig ...
type FilesTailsWatcherConfig struct {
	api                   *FilesTails
	filesPath             string
	filesMask             string
	relistFilesIntervalMS int64
	readFileIntervalMS    int64
	queueStringsSize      int64
}

// FilesTailsWatcher ...
type FilesTailsWatcher struct {
	api                   *FilesTails
	filesPath             string
	filesMask             string
	relistFilesIntervalMS int64
	readFileIntervalMS    int64
	receiver              InputStringReceiver

	input chan string
	files map[string]*fileTailWatcher

	ready sync.Mutex
}

// Constructor ...
func (filesTails FilesTails) Constructor(context context.Context, eventLoop jsEngine.EventLoop, runtime *goja.Runtime) {
	runtime.Set("FilesTails", &FilesTails{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// NewWatcher ...
func (filesTails *FilesTails) NewWatcher(filesMask string) *FilesTailsWatcherConfig {
	return &FilesTailsWatcherConfig{
		api:                   filesTails,
		filesMask:             filesMask,
		filesPath:             "",
		relistFilesIntervalMS: 1000,
		queueStringsSize:      256,
		readFileIntervalMS:    1000,
	}
}

// SetFilesPath ...
func (tailsWatcher *FilesTailsWatcherConfig) SetFilesPath(path string) *FilesTailsWatcherConfig {
	tailsWatcher.filesPath = path
	return tailsWatcher
}

// SetRelistFilesIntervalMS ...
func (tailsWatcher *FilesTailsWatcherConfig) SetRelistFilesIntervalMS(intervalMS int64) *FilesTailsWatcherConfig {
	tailsWatcher.relistFilesIntervalMS = intervalMS
	return tailsWatcher
}

// SetReadFileIntervalMS ...
func (tailsWatcher *FilesTailsWatcherConfig) SetReadFileIntervalMS(intervalMS int64) *FilesTailsWatcherConfig {
	tailsWatcher.readFileIntervalMS = intervalMS
	return tailsWatcher
}

// SetQueueStringsSize ...
func (tailsWatcher *FilesTailsWatcherConfig) SetQueueStringsSize(size int64) *FilesTailsWatcherConfig {
	tailsWatcher.queueStringsSize = size
	return tailsWatcher
}

// SendTo ...
func (tailsWatcher *FilesTailsWatcherConfig) SendTo(receiver InputStringReceiver) *FilesTailsWatcher {

	watcher := &FilesTailsWatcher{
		api:                   tailsWatcher.api,
		filesPath:             tailsWatcher.filesPath,
		filesMask:             tailsWatcher.filesMask,
		relistFilesIntervalMS: tailsWatcher.relistFilesIntervalMS,
		readFileIntervalMS:    tailsWatcher.readFileIntervalMS,
		receiver:              receiver,

		files: make(map[string]*fileTailWatcher),
		input: make(chan string, tailsWatcher.queueStringsSize),
	}

	_, err := receiver.getContext().NewContextFor(watcher, fmt.Sprintf("%v(%v)", watcher.filesPath, watcher.filesMask), "filesWatcher")
	if err != nil {
		log.Fatal(err)
	}

	return watcher
}

func (filesTailsWatcher *FilesTailsWatcher) appendNotExistingFilesToWatch(current context.Context, filesNames []string) {
	filesTailsWatcher.ready.Lock()

	for _, fileName := range filesNames {
		if _, found := filesTailsWatcher.files[fileName]; !found {

			watcher := newFileTailWatcher(filesTailsWatcher.filesPath, fileName, filesTailsWatcher.readFileIntervalMS, filesTailsWatcher.input, func() {
				filesTailsWatcher.deleteFileWatcher(fileName)
			})

			filesTailsWatcher.files[fileName] = watcher

			current.NewContextFor(watcher, fileName, "fileTailWatcher")
		}
	}

	filesTailsWatcher.ready.Unlock()
}

func (filesTailsWatcher *FilesTailsWatcher) deleteFileWatcher(fileName string) {
	filesTailsWatcher.ready.Lock()
	if _, found := filesTailsWatcher.files[fileName]; found {
		delete(filesTailsWatcher.files, fileName)
	}
	filesTailsWatcher.ready.Unlock()
}

// Go ...
func (filesTailsWatcher *FilesTailsWatcher) Go(current context.Context) {
loop:
	for {
		select {
		case <-time.After(time.Duration(filesTailsWatcher.relistFilesIntervalMS) * time.Millisecond):
			{ // query list of files that match to mask, and add them to dictionary
				fullFilesPath, err := filepath.Abs(filesTailsWatcher.filesPath)
				if err != nil {
					current.Log(err)
				} else {
					relativeFileNames, err := recursiveFilesSearch(fullFilesPath, fullFilesPath, filesTailsWatcher.filesMask)
					if err != nil {
						current.Log(err)
					} else {
						filesTailsWatcher.appendNotExistingFilesToWatch(current, relativeFileNames)
					}
				}
			}
			break
		case line := <-filesTailsWatcher.input:
			filesTailsWatcher.receiver.getInput() <- line
			/*
				jsLine := filesTailsWatcher.api.runtime.ToValue(line)
				_, err := filesTailsWatcher.api.eventLoop.CallHandler(filesTailsWatcher.parser, jsLine)
				if err != nil {
					current.Log(51, err.Error())
					current.Cancel()
				}
			*/
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}

func recursiveFilesSearch(rootPluginsPath string, currentFullPath string, filter string) ([]string, error) {
	result := []string{}

	files, err := ioutil.ReadDir(currentFullPath)
	if err != nil {
		return nil, err
	}

	path, err := filepath.Abs(currentFullPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			files, err := recursiveFilesSearch(rootPluginsPath, filepath.Join(path, file.Name()), filter)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		} else {
			match, _ := filepath.Match(filter, file.Name())
			if match {
				relativeName := strings.TrimPrefix(filepath.Join(path, file.Name()), rootPluginsPath)

				relativeNameWithoutSlash := relativeName
				if len(relativeNameWithoutSlash) > 0 {
					if relativeNameWithoutSlash[0] == 92 {
						relativeNameWithoutSlash = relativeNameWithoutSlash[1:]
					}
				}

				result = append(result, relativeNameWithoutSlash)
			}
		}
	}
	return result, nil
}
