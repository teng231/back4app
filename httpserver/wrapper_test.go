package httpserver

import (
	"log"
	"reflect"
	"testing"
)

func TestGetType(t *testing.T) {
	// return

	e := &Engine{}

	log.Print(reflect.TypeOf(e))
}
