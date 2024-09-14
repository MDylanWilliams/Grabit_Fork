package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComposeStatusString(t *testing.T) {
	line := composeStatusString(0, 1000, 0, 1, true, []string{"-", "\\", "|", "/", "-"}, 0, time.Now(), true)
	expected := "\x1b[33m\r-║░║          0/1 Complete          0B / 1,000B          0s Elapsed\x1b[0m"
	assert.Equal(t, expected, line)
}
