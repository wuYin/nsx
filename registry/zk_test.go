package registry

import (
	"fmt"
	"testing"
)

func TestZKRegistry(t *testing.T) {
	r := NewZKRegistry([]string{"127.0.0.1:2181"})

	err := r.Register("add-service", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}

	addrs, err := r.GetService("add-service")
	if err != nil || len(addrs) != 1 {
		t.Fatal(err, addrs)
	}
	fmt.Println(addrs)

	if err = r.UnRegister("add-service", "127.0.0.1:8080"); err != nil {
		t.Fatal(err)
	}
	if addrs, err := r.GetService("add-service"); err != nil || len(addrs) != 0 {
		t.Fatal(err, addrs)
	}
}
