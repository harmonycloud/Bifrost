package kube

import (
	"fmt"
	"testing"
)

func TestGetClien(t *testing.T) {
	client, _ := NewClient()
	fmt.Println(client)
}
