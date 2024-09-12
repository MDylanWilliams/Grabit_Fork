// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

var COMMENT_PREFIX = "//"

// Lock represents a grabit lockfile.
type Lock struct {
	path string
	conf config
}

type config struct {
	Resource []Resource
}

func NewLock(path string, newOk bool) (*Lock, error) {
	_, error := os.Stat(path)
	if os.IsNotExist(error) {
		if newOk {
			return &Lock{path: path}, nil
		} else {
			return nil, fmt.Errorf("file '%s' does not exist", path)
		}
	}
	var conf config
	file, err := os.Open(path)
	if err != nil {
		return nil, error
	}
	d := toml.NewDecoder(file)
	err = d.Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &Lock{path: path, conf: conf}, nil
}

func (l *Lock) AddResource(paths []string, algo string, tags []string, filename string) error {
	for _, u := range paths {
		if l.Contains(u) {
			return fmt.Errorf("resource '%s' is already present", u)
		}
	}
	r, err := NewResourceFromUrl(paths, algo, tags, filename)
	if err != nil {
		return err
	}
	l.conf.Resource = append(l.conf.Resource, *r)
	return nil
}

func (l *Lock) DeleteResource(path string) {
	newStatements := []Resource{}
	for _, r := range l.conf.Resource {
		if !r.Contains(path) {
			newStatements = append(newStatements, r)
		}
	}
	l.conf.Resource = newStatements
}

const NoFileMode = os.FileMode(0)

// strToFileMode converts a string to a os.FileMode.
func strToFileMode(perm string) (os.FileMode, error) {
	if perm == "" {
		return NoFileMode, nil
	}
	parsed, err := strconv.ParseUint(perm, 8, 32)
	if err != nil {
		return NoFileMode, err
	}
	return os.FileMode(parsed), nil
}

// Adds commas to number string at hundreds place, thousands place, etc.
// Ex: "12345678" -> "12,345,678"
func AddCommas(str string) string {
	for i := len(str) - 3; i >= 0; i -= 3 {
		str = str[:i] + "," + str[i:]
	}
	return str
}

// Download gets all the resources in this lock file and moves them to
// the destination directory.
func (l *Lock) Download(dir string, tags []string, notags []string, perm string, bar bool) error {
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		return fmt.Errorf("'%s' is not a directory", dir)
	}
	mode, err := strToFileMode(perm)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid permission definition", perm)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Filter in the resources that have all the required tags.
	tagFilteredResources := []Resource{}
	if len(tags) > 0 {
		for _, r := range l.conf.Resource {
			hasAllTags := true
			for _, tag := range tags {
				hasTag := false
				for _, rtag := range r.Tags {
					if tag == rtag {
						hasTag = true
						break
					}
				}
				if !hasTag {
					hasAllTags = false
					break
				}
			}
			if hasAllTags {
				tagFilteredResources = append(tagFilteredResources, r)
			}
		}
	} else {
		tagFilteredResources = l.conf.Resource
	}
	// Filter out the resources that have any 'notag' tag.
	filteredResources := []Resource{}
	if len(notags) > 0 {
		for _, r := range tagFilteredResources {
			hasTag := false
			for _, notag := range notags {
				for _, rtag := range r.Tags {
					if notag == rtag {
						hasTag = true
					}
				}
			}
			if !hasTag {
				filteredResources = append(filteredResources, r)
			}
		}
	} else {
		filteredResources = tagFilteredResources
	}

	total := len(filteredResources)
	if total == 0 {
		return fmt.Errorf("nothing to download")
	}
	errorCh := make(chan error, total)

	byteCh := make(chan int64)
	resourceSizes := make([]int64, 0, len(filteredResources))

	if bar {
		startTime := time.Now()

		//Loop through resources, fetching their metadata and totalling up their sizes in bytes.
		//This is not in a goroutine because a resource may finish downloading before its size has been calculated.
		//For now, we'll first calculate sizes, then begin downloading.
		fmt.Print(Color_Text("\rFetching file sizes...", "yellow"))

		var totalBytes int64 = 0
		for _, r := range filteredResources {
			resource := r
			httpClient := &http.Client{Timeout: 10 * time.Second}
			resp, err := httpClient.Head(resource.Urls[0])
			if err != nil {
				//errorCh <- err	We don't want a failed fetch to crash the whole program -- it is not crucial to know the total download size.
				//Instead, the print goroutine will check length of resourceSizes and notify the user by printing an error message in place of the byte counter.
				break
			}
			totalBytes += resp.ContentLength
			resourceSizes = append(resourceSizes, resp.ContentLength)
		}

		spinChars := [5]string{"-", "\\", "|", "/", "-"}
		spinI := 0 //Current char in spinChars.

		//SPINNER GOROUTINE.
		ticker := time.NewTicker(60 * time.Millisecond)
		go func() {
			//Continuously place garbage value (-1) in channel to prevent print goroutine from blocking and keep the spinner updating.
			for {
				select {
				case <-ticker.C:
					byteCh <- -1
				}
			}
		}()

		//PRINT GOROUTINE.
		go func() {
			resourcesDownloaded := 0
			var bytesDownloaded int64 = 0

			for {
				b := <-byteCh
				if b != -1 {
					resourcesDownloaded += 1
					bytesDownloaded += b
				}

				//Spinner loops through chars in spinChars to give the impression it is rotating.
				var spinner string
				if resourcesDownloaded < len(filteredResources) {
					spinner = spinChars[spinI]
					spinI += 1
					if spinI == len(spinChars) {
						spinI = 0
					}
				} else {
					spinner = "✔"
				}

				//Line is yellow while downloading, green when complete.
				var color string
				if resourcesDownloaded < len(filteredResources) {
					color = "yellow"
				} else {
					color = "green"
				}

				//Build progress bar string.
				barStr := "║"
				for i := 0; i < resourcesDownloaded; i += 1 {
					barStr += "█"
				}

				if resourcesDownloaded < len(filteredResources) {
					barStr += "░"
				}

				for i := resourcesDownloaded + 1; i < len(filteredResources); i += 1 {
					barStr += "_"
				}
				barStr += "║"

				completeStr := strconv.Itoa(resourcesDownloaded) + "/" + strconv.Itoa(len(filteredResources)) + " Complete"

				var byteStr string
				if len(resourceSizes) == len(filteredResources) {
					byteStr = AddCommas(strconv.Itoa(int(bytesDownloaded))) + "B / " + byteStr + AddCommas(strconv.Itoa(int(totalBytes))) + "B"
				} else {
					byteStr = "<ERROR FETCHING RESOURCE SIZES>"
				}

				elapsedStr := strconv.Itoa(int(time.Now().Sub(startTime).Round(time.Second).Seconds())) + "s Elapsed"

				//Build and print line.
				pad := "          "
				line := "\r" + spinner + barStr + pad + completeStr + pad + byteStr + pad + elapsedStr //"\r" lets us clear the line.

				fmt.Print(Color_Text(line, color))

				if resourcesDownloaded == len(filteredResources) {
					fmt.Println()
					break
				}
			}
		}()
	}
	for i, r := range filteredResources {
		resource := r
		go func() {

			err := resource.Download(dir, mode, ctx)

			errorCh <- err

			if len(resourceSizes) == len(filteredResources) {
				byteCh <- resourceSizes[i]
			} else {
				byteCh <- 0
			}
		}()
	}
	done := 0
	for err := range errorCh {
		if err != nil {
			return err
		}
		done += 1
		if done == total {
			return nil
		}
	}

	return nil
}

// Save this lock file to disk.
func (l *Lock) Save() error {
	res, err := toml.Marshal(l.conf)
	if err != nil {
		return err
	}
	file, err := os.Create(l.path)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = w.Write(res)
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}

// Contains returns true if this lock file contains the
// given resource url.
func (l *Lock) Contains(url string) bool {
	for _, r := range l.conf.Resource {
		for _, u := range r.Urls {
			if url == u {
				return true
			}
		}
	}
	return false
}
