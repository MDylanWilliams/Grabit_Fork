package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

func TestSLSpinner(t *testing.T) {
	resources := createResources(1, t)
	ctx, _ := context.WithCancel(context.Background())
	st := NewStatusLine(ctx, &resources)
	err := st.InitResourcesSizes()
	assert.Nil(t, err)

	st.Start(true)
	assert.Equal(t, "\r-[      ]          0/1 Resources          0 B / 6 B          0s elapsed", st.GetStatusString())
	time.Sleep(60 * time.Millisecond) // Give extra ms to allow SL to update.
	assert.Equal(t, "\r\\[      ]          0/1 Resources          0 B / 6 B          0s elapsed", st.GetStatusString())
	time.Sleep(60 * time.Millisecond)
	assert.Equal(t, "\r|[      ]          0/1 Resources          0 B / 6 B          0s elapsed", st.GetStatusString())
	time.Sleep(60 * time.Millisecond)
	assert.Equal(t, "\r/[      ]          0/1 Resources          0 B / 6 B          0s elapsed", st.GetStatusString())
	time.Sleep(60 * time.Millisecond)
	assert.Equal(t, "\r-[      ]          0/1 Resources          0 B / 6 B          0s elapsed", st.GetStatusString())
}

func TestSLWith2Resources(t *testing.T) {
	resources := createResources(2, t)
	ctx, _ := context.WithCancel(context.Background())
	st := NewStatusLine(ctx, &resources)
	err := st.InitResourcesSizes()
	assert.Nil(t, err)

	st.Start(false)
	assert.Equal(t, "\r-[            ]          0/2 Resources          0 B / 12 B          0s elapsed", st.GetStatusString())
	st.Increment(0)
	assert.Equal(t, "\r-[██████      ]          1/2 Resources          6 B / 12 B          0s elapsed", st.GetStatusString())
	st.Increment(1)
	assert.Equal(t, "\r✔[████████████]          2/2 Resources          12 B / 12 B          0s elapsed", st.GetStatusString())

}

func TestSLWith1000Resources(t *testing.T) {
	resources := createResources(1000, t)
	ctx, _ := context.WithCancel(context.Background())
	st := NewStatusLine(ctx, &resources)
	err := st.InitResourcesSizes()
	assert.Nil(t, err)

	st.Start(false)
	assert.Equal(t, "\r-[                    ]          0/1000 Resources          0 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(0)
	assert.Equal(t, "\r-[                    ]          1/1000 Resources          6 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(1)
	assert.Equal(t, "\r-[                    ]          2/1000 Resources          12 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(2)
	assert.Equal(t, "\r-[                    ]          3/1000 Resources          18 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(3)
	assert.Equal(t, "\r-[                    ]          4/1000 Resources          24 B / 6.0 kB          0s elapsed", st.GetStatusString())
	for i := 4; i < 1000; i++ {
		st.Increment(i)
	}
	assert.Equal(t, st.GetStatusString(), "\r✔[████████████████████]          1000/1000 Resources          6.0 kB / 6.0 kB          0s elapsed")
}

func TestSLWithCancelledContext(t *testing.T) {
	resources := createResources(2, t)
	ctx, cancel := context.WithCancel(context.Background())
	st := NewStatusLine(ctx, &resources)
	err := st.InitResourcesSizes()
	assert.Nil(t, err)

	assert.Equal(t, false, st.isRunning.Load())
	st.Start(false)
	st.Increment(0)
	cancel()
	time.Sleep(10 * time.Millisecond) // Give SL time to Stop.
	assert.NotEqual(t, true, st.isRunning.Load())
}

func createResources(num int, t *testing.T) []Resource {
	content := `abcdef`
	port := test.TestHttpHandler(content, t)
	resources := []Resource{}
	algo := "sha256"
	for i := 0; i < num; i++ {
		resource := Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test%d.html", port, i)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: ""}
		resources = append(resources, resource)
	}
	return resources
}
