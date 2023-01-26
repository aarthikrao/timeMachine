package jobmodels

import (
	"bytes"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestTriggerTimeBytes(t *testing.T) {
	var j Job
	btime := j.TriggerTimeBytes()
	if !bytes.Equal(btime, []byte("0")) {
		t.Fail()
		t.Error("Wrong encoding")
	}
	j.TriggerTime = 20
	btime = j.TriggerTimeBytes()
	if !bytes.Equal(btime, []byte("20")) {
		t.Fail()
		t.Error("Wrong encoding")
	}
	rand.Seed(time.Now().Unix())
	j.TriggerTime = rand.Int63()
	btime = j.TriggerTimeBytes()
	if !bytes.Equal(btime, []byte(strconv.FormatInt(j.TriggerTime, 10))) {
		t.Fail()
		t.Error("Wrong encoding of random number")
	}
}

func TestGetMinuteBucketName(t *testing.T) {
	var j Job
	bucketName := j.GetMinuteBucketName()
	if !bytes.Equal([]byte("0"), bucketName) {
		t.Fail()
		t.Error("nil job should return \"0\" as bucketName")
	}
	bucketName = j.GetMinuteBucketName()
	if !bytes.Equal([]byte("0"), bucketName) {
		t.Fail()
		t.Error("job with negative triggertime should return \"0\" as bucketName")
	}
	now := time.Now()
	j.TriggerTime = now.UnixMilli()
	bucketName = j.GetMinuteBucketName()
	minutesSinceEppoch := now.Unix() / int64(60)
	expected := []byte(strconv.FormatInt(minutesSinceEppoch, 10))
	if !bytes.Equal(expected, bucketName) {
		t.Fail()
		t.Errorf("invalid minuteBucketName for given time, expected=%v found=%v", string(expected), string(bucketName))
	}
}

func TestGetUniqueKey(t *testing.T) {
	var j Job
	id := j.GetUniqueKey("")
	if id != nil {
		t.Fail()
		t.Error("nil collections shouldn't be allowed")
	}
	id = j.GetUniqueKey("test")
	if !bytes.Equal(id, []byte("test_0")) {
		t.Fail()
		t.Error("not working")
	}
	rand.Seed(time.Now().Unix())
	randN := strconv.Itoa(rand.Int())
	j.TriggerTime = rand.Int63()
	id = j.GetUniqueKey(randN)
	if !bytes.Equal(id, []byte(randN+"_"+strconv.FormatInt(j.TriggerTime, 10))) {
		t.Fail()
		t.Error("unique key not working")
	}
}
