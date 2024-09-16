package internal

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Status_Line struct {
	resources []Resource
	indexCh   chan int
}

func (st *Status_Line) run() {
	startTime := time.Now()

	resourceSizes, totalBytes, sizingSuccess := getResourcesSizes(st, 10000)
	startTicker(st)

	var bytesDownloaded int64
	resourcesDownloaded := 0
	go func() {
		spinChars := [5]string{"-", "\\", "|", "/", "-"}
		spinI := 0

		for {
			if i := <-st.indexCh; i != -1 {
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

func getResourcesSizes(st *Status_Line, timeoutMs int) ([]int64, int64, bool) {
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

func startTicker(st *Status_Line) {
	ticker := time.NewTicker(50 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				st.indexCh <- -1
			}
		}
	}()
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

	completeStr := strconv.Itoa(resourcesDownloaded) + "/" + strconv.Itoa(numResources) + " Complete"

	var byteStr string
	if sizingSuccess {
		byteStr = AddCommas(strconv.Itoa(int(bytesDownloaded))) + "B / " + byteStr + AddCommas(strconv.Itoa(int(totalBytes))) + "B"
	} else {
		byteStr = "<issue_fetching_resource_sizes>"
	}

	elapsedStr := strconv.Itoa(int(time.Since(startTime).Round(time.Second).Seconds())) + "s Elapsed"

	pad := "          "
	line := "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr //"\r" lets us clear the line.

	return line
}

// Adds commas to number string at hundreds place, thousands place, etc.
// Ex: "12345678" -> "12,345,678"
func AddCommas(str string) string {
	for i := len(str) - 3; i >= 0; i -= 3 {
		str = str[:i] + "," + str[i:]
	}
	return str
}
