package grid

import (
	"os"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	actualBytes, err := New(42, 44, 46)
	if err != nil {
		t.Fatal(err)
	}
	expectedBytes, err := os.ReadFile("./testdata/z42-x44-y46.png")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actualBytes, expectedBytes) {
		t.Fatal("actualBytes != expectedBytes")
	}
}
