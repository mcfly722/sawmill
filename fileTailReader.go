package main

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/mcfly722/goPackages/context"
)

type tailFileReader struct {
	input                         chan string
	fullFileName                  string
	intervalBetweenFileReadingsMS int64
}

func newTailFileReader(fullFileName string, intervalBetweenFileReadingsMS int64) *tailFileReader {
	return &tailFileReader{
		input:                         make(chan string, 100),
		fullFileName:                  fullFileName,
		intervalBetweenFileReadingsMS: intervalBetweenFileReadingsMS,
	}
}

func (tailFileReader *tailFileReader) init(parentContext context.Context) chan string {
	parentContext.NewContextFor(tailFileReader, tailFileReader.fullFileName, "tailFileReader")
	return tailFileReader.input
}

// Go ...
func (tailFileReader *tailFileReader) Go(current context.Context) {
	lastOffset := int64(0)

	duration := time.Duration(rand.Int63n(tailFileReader.intervalBetweenFileReadingsMS)) * time.Millisecond

loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval
			{
				duration = time.Duration(tailFileReader.intervalBetweenFileReadingsMS) * time.Millisecond // after first start we change interval dutation to required

				file, err := os.Open(tailFileReader.fullFileName)
				if err != nil {
					lastOffset = int64(0)
					break
				}
				//fmt.Printf(fmt.Sprintf(">%v", lastOffset))

				{

					{ // set read possition to last one
						fileStat, err := file.Stat()
						if err != nil || lastOffset > fileStat.Size() {
							lastOffset = int64(0)
							break
						}

						_, err = file.Seek(lastOffset, io.SeekStart)
						if err != nil {
							//							current.Log(err)
							lastOffset = int64(0)
						}
					}

					{ // read to end
						scanner := bufio.NewScanner(file)
						scanner.Split(bufio.ScanLines)

						for scanner.Scan() {
							tailFileReader.input <- scanner.Text()
						}
					}

					{ // save last read position
						pos, err := file.Seek(0, io.SeekCurrent)
						if err == nil {
							lastOffset = pos
						}
					}

				}

				//fmt.Printf("<")
				file.Close()

				break
			}
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}
