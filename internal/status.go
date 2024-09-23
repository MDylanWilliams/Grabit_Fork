package internal

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

type StatusLine struct {
	resources []Resource

	resourceSizes       []int64
	totalBytes          int64
	sizingSuccess       bool
	bytesDownloaded     int64
	resourcesDownloaded int
	spinI               int
	indexCh             chan int
	startTime           time.Time
}

var spinChars = [5]string{"-", "\\", "|", "/", "-"}

// Create, initialize, and start a new StatusLine.
func NewStatusLine(resources []Resource) *StatusLine {
	st := StatusLine{}
	st.resources = resources
	st.indexCh = make(chan int)
	st.resourceSizes, st.totalBytes, st.sizingSuccess = getResourcesSizes(resources, 10000)

	st.start()

	return &st
}

// Inform StatusLine that resource has finished downloading.
func (st *StatusLine) increment(i int) {
	st.indexCh <- i
}

// Begin goroutine and for loop.
func (st *StatusLine) start() {
	st.startTime = time.Now()

	go func() {
		st.spinI = 0

		for {

			//Block until value is inserted into indexCh (>= 0 when resource done downloading, -1 every 50 milliseconds to keep timer and spinner updating).
			var i int
			select {
			case i = <-st.indexCh:
			case <-time.After(50 * time.Millisecond):
				i = -1
			}
			if i != -1 {
				st.bytesDownloaded += st.resourceSizes[i]
				st.resourcesDownloaded++
			}

			//Update spinner.
			st.spinI += 1
			if st.spinI == len(spinChars) {
				st.spinI = 0
			}

			// Print line.
			anyRemaining := st.resourcesDownloaded < len(st.resources)
			line := composeStatusString(st.bytesDownloaded, st.totalBytes, st.resourcesDownloaded, len(st.resources), st.sizingSuccess, spinChars[:], st.spinI, st.startTime, anyRemaining, 20)
			fmt.Print(line)

			//Exit if done.
			if !anyRemaining {
				fmt.Println()
				break
			}

		}
	}()

}

func getResourcesSizes(resources []Resource, timeoutMs int) ([]int64, int64, bool) {
	fmt.Print("\rFetching resource sizes...")
	resourceSizes := make([]int64, len(resources))
	for i := 0; i < len(resourceSizes); i++ {
		resourceSizes[i] = 0
	}

	var totalBytes int64 = 0
	sizingSuccess := true
	for i, r := range resources {
		resource := r
		httpClient := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
		resp, err := httpClient.Head(resource.Urls[0])
		if err != nil {
			sizingSuccess = false
			break
		}
		totalBytes += resp.ContentLength
		resourceSizes[i] = resp.ContentLength
	}

	return resourceSizes, totalBytes, sizingSuccess
}

func composeStatusString(bytesDownloaded int64, totalBytes int64, resourcesDownloaded int, numResources int, sizingSuccess bool, spinChars []string, spinI int, startTime time.Time, anyRemaining bool, barLength int) string {
	var spinner string
	if anyRemaining {
		spinner = spinChars[spinI]
	} else {
		spinner = "✔"
	}

	barStr := "["
	squareSize := totalBytes / int64(barLength)
	for i := 0; i < int(bytesDownloaded/squareSize); i += 1 {
		barStr += "█"
	}
	if resourcesDownloaded < numResources {
		barStr += " "
	}
	for i := int(bytesDownloaded/squareSize) + 1; i < 20; i += 1 {
		barStr += " "
	}
	barStr += "]"

	completeStr := strconv.Itoa(resourcesDownloaded) + "/" + strconv.Itoa(numResources) + " Resources"

	var byteStr string
	if sizingSuccess {
		byteStr = humanize.Bytes(uint64(bytesDownloaded)) + " / " + humanize.Bytes(uint64(totalBytes))
	} else {
		byteStr = "<issue_fetching_resource_sizes>"
	}

	elapsedStr := strconv.Itoa(int(time.Since(startTime).Round(time.Second).Seconds())) + "s Elapsed"

	pad := "          "
	line := "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr // "\r" lets us clear the line.

	return line
}
