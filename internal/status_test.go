package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

func TestSLWith2Resources(t *testing.T) {
	// Create and init new SL with generate resources.
	resources := createResources(2, t)
	ctx, _ := context.WithCancel(context.Background())
	st1 := NewStatusLine(ctx, &resources)
	err := st1.InitResourcesSizes()
	assert.Nil(t, err)

	st1.Start(false)
	assert.Equal(t, "\r-[            ]          0/2 Resources          0 B / 12 B          0s elapsed", st1.GetStatusString())
	st1.Increment(0)
	assert.Equal(t, "\r-[██████      ]          1/2 Resources          6 B / 12 B          0s elapsed", st1.GetStatusString())
	st1.Increment(1)
	assert.Equal(t, "\r✔[████████████]          2/2 Resources          12 B / 12 B          0s elapsed", st1.GetStatusString())

}

func TestSLWith1000Resources(t *testing.T) {
	resources := createResources(1000, t)
	ctx, _ := context.WithCancel(context.Background())
	st2 := NewStatusLine(ctx, &resources)
	err := st2.InitResourcesSizes()
	assert.Nil(t, err)

	st2.Start(false)
	assert.Equal(t, "\r-[                    ]          0/1000 Resources          0 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(0)
	assert.Equal(t, "\r-[                    ]          1/1000 Resources          6 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(1)
	assert.Equal(t, "\r-[                    ]          2/1000 Resources          12 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(2)
	assert.Equal(t, "\r-[                    ]          3/1000 Resources          18 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(3)
	assert.Equal(t, "\r-[                    ]          4/1000 Resources          24 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	for i := 4; i < 1000; i++ {
		st2.Increment(i)
	}
	assert.Equal(t, st2.GetStatusString(), "\r✔[████████████████████]          1000/1000 Resources          6.0 kB / 6.0 kB          0s elapsed")
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
	time.Sleep(1 * time.Millisecond) // Give SL time to Stop.
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
