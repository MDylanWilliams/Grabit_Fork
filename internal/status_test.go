package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComposeStatusString(t *testing.T) {
	line := composeStatusString(0, 1000, 0, 1, true, []string{"-", "\\", "|", "/", "-"}, 0, time.Now(), true)
	expected := "\r-[ ]          0/1 Complete          0B / 1,000B          0s Elapsed"
	assert.Equal(t, expected, line)
}
