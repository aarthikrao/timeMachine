package datastore

import (
	"os"
	"testing"
)

func TestCreateBoltDataStore(t *testing.T) {
	tempDir := os.TempDir()
	if len(tempDir) == 0 {
		t.Error("Couldn't find temporary directory")
		t.Fail()
		return
	}

	tempFile, err := os.CreateTemp(tempDir, "_bolt_temp")
	if err != nil {
		t.Errorf("couldn't create temporary file %v", err)
		t.Fail()
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	dbStore, err := CreateBoltDataStore(tempFile.Name())
	if err != nil {
		t.Errorf("error while opening the db %v", err)
		t.Fail()
		return
	}
	defer dbStore.Close()

}
