package timeutils

import (
	"testing"
	"time"
)

func TestGetCurrentMillis(t *testing.T) {
	millis := GetCurrentMillis()
	nowMillis := time.Now().UnixMilli()
	if nowMillis != millis {
		t.Fail()
		t.Error("invalid number of milliseconds")
	}

}
