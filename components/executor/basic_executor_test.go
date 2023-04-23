package executor

import (
	"testing"
	"time"
)

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now().Unix()
	}
}

func BenchmarkCustomTimer(b *testing.B) {
	var dis dispatcher
	go dis.startTimer()
	for i := 0; i < b.N; i++ {
		dis.getCurrentTime()
	}
}
