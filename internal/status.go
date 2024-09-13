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

	//Get file sizes.
	fmt.Print(ColorText("\rFetching resource sizes...", "yellow"))
	resourceSizes := make([]int64, len(st.resources))
	for i := 0; i < len(resourceSizes); i++ {
		resourceSizes[i] = 0
	}
	var totalBytes int64 = 0
	sizingSuccess := true
	for i, r := range st.resources {
		resource := r
		httpClient := &http.Client{Timeout: 10 * time.Second}
		resp, err := httpClient.Head(resource.Urls[0])
		if err != nil {
			sizingSuccess = false
			break
		}
		totalBytes += resp.ContentLength
		resourceSizes[i] = resp.ContentLength
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				st.indexCh <- -1
			}
		}
	}()

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

			var spinner string
			if anyRemaining {
				spinner = spinChars[spinI]
				spinI += 1
				if spinI == len(spinChars) {
					spinI = 0
				}
			} else {
				spinner = "✔"
			}

			barStr := "║"
			for i := 0; i < resourcesDownloaded; i += 1 {
				barStr += "█"
			}
			if resourcesDownloaded < len(st.resources) {
				barStr += "░"
			}
			for i := resourcesDownloaded + 1; i < len(st.resources); i += 1 {
				barStr += "_"
			}
			barStr += "║"

			completeStr := strconv.Itoa(resourcesDownloaded) + "/" + strconv.Itoa(len(st.resources)) + " Complete"

			var byteStr string
			if sizingSuccess {
				byteStr = AddCommas(strconv.Itoa(int(bytesDownloaded))) + "B / " + byteStr + AddCommas(strconv.Itoa(int(totalBytes))) + "B"
			} else {
				byteStr = "<ISSUE FETCHING RESOURCE SIZES>"
			}

			elapsedStr := strconv.Itoa(int(time.Since(startTime).Round(time.Second).Seconds())) + "s Elapsed"

			var color string
			if anyRemaining {
				color = "yellow"
			} else {
				color = "green"
			}

			pad := "          "
			line := "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr //"\r" lets us clear the line.
			fmt.Print(ColorText(line, color))

			if !anyRemaining {
				fmt.Println()
				break
			}

		}
	}()

}
