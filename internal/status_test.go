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
	err := st1.InitResourcesSizes(1)
	assert.NotNil(t, err)
	err = st1.InitResourcesSizes(1000)
	assert.Nil(t, err)

	st1.Start(false)
	assert.Equal(t, "\r-[            ]          0/2 Resources          0 B / 12 B          0s elapsed", st1.GetStatusString())
	st1.Increment(0)
	time.Sleep(10 * time.Millisecond) // Give SL loop time to update fields.
	assert.Equal(t, "\r\\[██████      ]          1/2 Resources          6 B / 12 B          0s elapsed", st1.GetStatusString())
	st1.Increment(1)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r✔[████████████]          2/2 Resources          12 B / 12 B          0s elapsed", st1.GetStatusString())

}

func TestSLWith1000Resources(t *testing.T) {
	resources := createResources(1000, t)
	ctx, _ := context.WithCancel(context.Background())
	st2 := NewStatusLine(ctx, &resources)
	err := st2.InitResourcesSizes(1)
	assert.NotNil(t, err)
	err = st2.InitResourcesSizes(1000)
	assert.Nil(t, err)

	st2.Start(false)
	assert.Equal(t, "\r-[                    ]          0/1000 Resources          0 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(0)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r\\[                    ]          1/1000 Resources          6 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(1)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r|[                    ]          2/1000 Resources          12 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(2)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r/[                    ]          3/1000 Resources          18 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	st2.Increment(3)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r-[                    ]          4/1000 Resources          24 B / 6.0 kB          0s elapsed", st2.GetStatusString())
	for i := 4; i < 1000; i++ {
		st2.Increment(i)
	}
	assert.Equal(t, "\r✔[████████████████████]          1000/1000 Resources          6.0 kB / 6.0 kB          0s elapsed", st2.GetStatusString())
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
