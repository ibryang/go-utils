package example_test

import (
	"testing"

	"github.com/ibryang/go-utils/changedpi"
)

func TestChangeDpi(t *testing.T) {
	output, err := changedpi.ChangeDpiByPath("/Users/ibryang/Downloads/笔记本月份花图案/棕色/棕色花2.png", 300)
	if err != nil {
		t.Error(err)
	}
	err = changedpi.SaveImage("example300.jpeg", output)
	if err != nil {
		t.Error(err)
	}
}
