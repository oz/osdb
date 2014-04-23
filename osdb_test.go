package osdb

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHashWithNonExistentFile(t *testing.T) {
	_, err := Hash("./nonexistent")
	if err == nil {
		t.Fatalf("Expected an error, got none")

	}
}

func TestHashWithTooSmallFile(t *testing.T) {
	err := ioutil.WriteFile("./small", []byte("too small"), 0644)
	defer os.Remove("./small")
	if err != nil {
		t.Fatalf("Can't create small file")
	}

	_, err = Hash("./small")
	if err == nil {
		t.Fatalf("Expected an error, got none")
	}
}

func TestHashWithSample(t *testing.T) {
	data := make([]byte, ChunkSize*2)
	var dataHash uint64 = 0x6c62616c62636cc3
	copy(data, []byte("blablabla"))

	err := ioutil.WriteFile("./sample", data, 0644)
	defer os.Remove("./sample")
	if err != nil {
		t.Fatalf("Can't create sample file")
	}

	hash, err := Hash("./sample")
	if err != nil {
		t.Fatalf("Expected hash, got error: %v", err)
	}
	if hash != dataHash {
		t.Fatalf("Expected hash 0x%016x, got 0x%016x", dataHash, hash)
	}
}

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	if err != nil {
		t.Fatalf("Can't allocate new client: %v", err)
	}
}
