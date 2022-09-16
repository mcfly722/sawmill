package main

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// FileTailWatcher ...
type fileTailWatcher struct {
	fileName           string
	filePath           string
	readFileIntervalMS int64
	input              chan string
	onDispose          func()
}

func newFileTailWatcher(filePath string, fileName string, readFileIntervalMS int64, input chan string, onDispose func()) *fileTailWatcher {
	return &fileTailWatcher{
		fileName:           fileName,
		filePath:           filePath,
		readFileIntervalMS: readFileIntervalMS,
		input:              input,
		onDispose:          onDispose,
	}
}

func continueToReadFileByStrings(filePath string, fileName string, lastOffset int64, pusher func(string)) (int64, error) {

	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return -1, err
	}

	fullFileName := filepath.Join(fullPath, fileName)

	// open file
	file, err := os.Open(fullFileName)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	{ // set read possition to last one
		fileStat, err := file.Stat()
		if err != nil || lastOffset > fileStat.Size() {
			return -1, err
		}

		_, err = file.Seek(lastOffset, io.SeekStart)
		if err != nil {
			return -1, err
		}
	}

	{ // read to end
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			pusher(scanner.Text())
		}
	}

	// save last read position and exit
	return file.Seek(0, io.SeekCurrent)
}

// Go ...
func (fileTailWatcher *fileTailWatcher) Go(current context.Context) {
	lastOffset := int64(0)

	duration := time.Duration(rand.Int63n(fileTailWatcher.readFileIntervalMS)) * time.Millisecond

loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval
			duration = time.Duration(fileTailWatcher.readFileIntervalMS) * time.Millisecond // after first start we change interval dutation to required

			newOffset, err := continueToReadFileByStrings(fileTailWatcher.filePath, fileTailWatcher.fileName, lastOffset, func(line string) {
				fileTailWatcher.input <- line
			})

			if err != nil {
				current.Log(120, err)
				break loop
			}

			lastOffset = newOffset

			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}

	fileTailWatcher.onDispose()
}
