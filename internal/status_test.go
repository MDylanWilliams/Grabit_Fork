package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

func TestGetStatusString(t *testing.T) {
	// Test 1000 resources.
	content := `abcdef`
	port := test.TestHttpHandler(content, t)
	resources := []Resource{}
	algo := "sha256"
	for i := 0; i < 1000; i++ {
		resource := Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test%d.html", port, i)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: ""}
		resources = append(resources, resource)
	}
	ctx, _ := context.WithCancel(context.Background())

	// Test resource generation.
	assert.Len(t, resources, 1000)

	// Test StatusLine initialization and initResourcesSizes().
	st, err := NewStatusLine(ctx, &resources)
	assert.Nil(t, err)

	// Test GetStatusString() and Increment().
	st.Start(false)
	assert.Equal(t, "\r-[                    ]          0/1000 Resources          0 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(0)
	time.Sleep(10 * time.Millisecond) // Give StatusLine loop time to update fields.
	assert.Equal(t, "\r\\[                    ]          1/1000 Resources          6 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(1)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r|[                    ]          2/1000 Resources          12 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(2)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r/[                    ]          3/1000 Resources          18 B / 6.0 kB          0s elapsed", st.GetStatusString())
	st.Increment(3)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "\r-[                    ]          4/1000 Resources          24 B / 6.0 kB          0s elapsed", st.GetStatusString())
}
