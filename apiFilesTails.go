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

// FilesTailsWatcher ...
type FilesTailsWatcher struct {
	api                   *FilesTails
	filesPath             string
	filesMask             string
	relistFilesIntervalMS int64
	readFileIntervalMS    int64
	queueStringsSize      int64
}

// StartedFilesTailsWatcher ...
type StartedFilesTailsWatcher struct {
	api                   *FilesTails
	filesPath             string
	filesMask             string
	relistFilesIntervalMS int64
	readFileIntervalMS    int64
	parser                *goja.Callable

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
func (filesTails *FilesTails) NewWatcher(filesMask string) *FilesTailsWatcher {
	return &FilesTailsWatcher{
		api:                   filesTails,
		filesMask:             filesMask,
		filesPath:             "",
		relistFilesIntervalMS: 1000,
		queueStringsSize:      256,
		readFileIntervalMS:    1000,
	}
}

// SetFilesPath ...
func (tailsWatcher *FilesTailsWatcher) SetFilesPath(path string) *FilesTailsWatcher {
	tailsWatcher.filesPath = path
	return tailsWatcher
}

// SetRelistFilesIntervalMS ...
func (tailsWatcher *FilesTailsWatcher) SetRelistFilesIntervalMS(intervalMS int64) *FilesTailsWatcher {
	tailsWatcher.relistFilesIntervalMS = intervalMS
	return tailsWatcher
}

// SetReadFileIntervalMS ...
func (tailsWatcher *FilesTailsWatcher) SetReadFileIntervalMS(intervalMS int64) *FilesTailsWatcher {
	tailsWatcher.readFileIntervalMS = intervalMS
	return tailsWatcher
}

// SetQueueStringsSize ...
func (tailsWatcher *FilesTailsWatcher) SetQueueStringsSize(size int64) *FilesTailsWatcher {
	tailsWatcher.queueStringsSize = size
	return tailsWatcher
}

// StartWithParser ...
func (tailsWatcher *FilesTailsWatcher) StartWithParser(parser *goja.Callable) *StartedFilesTailsWatcher {

	startedWatcher := &StartedFilesTailsWatcher{
		api:                   tailsWatcher.api,
		filesPath:             tailsWatcher.filesPath,
		filesMask:             tailsWatcher.filesMask,
		relistFilesIntervalMS: tailsWatcher.relistFilesIntervalMS,
		readFileIntervalMS:    tailsWatcher.readFileIntervalMS,
		parser:                parser,

		files: make(map[string]*fileTailWatcher),
		input: make(chan string, tailsWatcher.queueStringsSize),
	}

	_, err := tailsWatcher.api.context.NewContextFor(startedWatcher, fmt.Sprintf("%v(%v)", startedWatcher.filesPath, startedWatcher.filesMask), "filesWatcher")
	if err != nil {
		log.Fatal(err)
	}

	return startedWatcher
}

func (startedFilesTailsWatcher *StartedFilesTailsWatcher) appendNotExistingFilesToWatch(current context.Context, filesNames []string) {
	startedFilesTailsWatcher.ready.Lock()

	for _, fileName := range filesNames {
		if _, found := startedFilesTailsWatcher.files[fileName]; !found {

			watcher := newFileTailWatcher(startedFilesTailsWatcher.filesPath, fileName, startedFilesTailsWatcher.readFileIntervalMS, startedFilesTailsWatcher.input, func() {
				startedFilesTailsWatcher.deleteFileWatcher(fileName)
			})

			startedFilesTailsWatcher.files[fileName] = watcher

			current.NewContextFor(watcher, fileName, "fileTailWatcher")
		}
	}

	startedFilesTailsWatcher.ready.Unlock()
}

func (startedFilesTailsWatcher *StartedFilesTailsWatcher) deleteFileWatcher(fileName string) {
	startedFilesTailsWatcher.ready.Lock()
	if _, found := startedFilesTailsWatcher.files[fileName]; found {
		delete(startedFilesTailsWatcher.files, fileName)
	}
	startedFilesTailsWatcher.ready.Unlock()
}

// Go ...
func (startedFilesTailsWatcher *StartedFilesTailsWatcher) Go(current context.Context) {
loop:
	for {
		select {
		case <-time.After(time.Duration(startedFilesTailsWatcher.relistFilesIntervalMS) * time.Millisecond):
			{ // query list of files that match to mask, and add them to dictionary
				fullFilesPath, err := filepath.Abs(startedFilesTailsWatcher.filesPath)
				if err != nil {
					current.Log(err)
				} else {
					relativeFileNames, err := recursiveFilesSearch(fullFilesPath, fullFilesPath, startedFilesTailsWatcher.filesMask)
					if err != nil {
						current.Log(err)
					} else {
						startedFilesTailsWatcher.appendNotExistingFilesToWatch(current, relativeFileNames)
					}
				}
			}
			break
		case line := <-startedFilesTailsWatcher.input:
			jsLine := startedFilesTailsWatcher.api.runtime.ToValue(line)
			_, err := startedFilesTailsWatcher.api.eventLoop.CallHandler(startedFilesTailsWatcher.parser, jsLine)
			if err != nil {
				current.Log(51, err.Error())
				current.Cancel()
			}

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
