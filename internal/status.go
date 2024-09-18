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
	indexCh   chan int
}

var spinChars = [5]string{"-", "\\", "|", "/", "-"}

func (st *StatusLine) run() {
	startTime := time.Now()

	resourceSizes, totalBytes, sizingSuccess := getResourcesSizes(st, 10000)

	var bytesDownloaded int64
	resourcesDownloaded := 0
	go func() {
		spinI := 0

		for {
			var i int
			select {
			case i = <-st.indexCh:
			case <-time.After(50 * time.Millisecond):
				i = -1
			}
			if i != -1 {
				bytesDownloaded += resourceSizes[i]
				resourcesDownloaded++
			}
			anyRemaining := resourcesDownloaded < len(st.resources)

			spinI += 1
			if spinI == len(spinChars) {
				spinI = 0
			}

			line := composeStatusString(bytesDownloaded, totalBytes, resourcesDownloaded, len(st.resources), sizingSuccess, spinChars[:], spinI, startTime, anyRemaining)
			fmt.Print(line)
			if !anyRemaining {
				fmt.Println()
				break
			}

		}
	}()

}

func getResourcesSizes(st *StatusLine, timeoutMs int) ([]int64, int64, bool) {
	fmt.Print("\rFetching resource sizes...")
	resourceSizes := make([]int64, len(st.resources))
	for i := 0; i < len(resourceSizes); i++ {
		resourceSizes[i] = 0
	}
	var totalBytes int64 = 0
	sizingSuccess := true
	for i, r := range st.resources {
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

func composeStatusString(bytesDownloaded int64, totalBytes int64, resourcesDownloaded int, numResources int, sizingSuccess bool, spinChars []string, spinI int, startTime time.Time, anyRemaining bool) string {
	var spinner string
	if anyRemaining {
		spinner = spinChars[spinI]
	} else {
		spinner = "✔"
	}

	barStr := "["
	for i := 0; i < resourcesDownloaded; i += 1 {
		barStr += "█"
	}
	if resourcesDownloaded < numResources {
		barStr += " "
	}
	for i := resourcesDownloaded + 1; i < numResources; i += 1 {
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
