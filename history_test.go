package investing

import (
	"testing"
	"time"
)

func TestGetHistory(t *testing.T) {
	start := time.Date(2018, 6, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2018, 6, 12, 0, 0, 0, 0, time.UTC)
	lines, err := GetHistory("38013", "D", start, end)
	if err != nil {
		t.Error(err)
	}
	t.Log(lines)
}
