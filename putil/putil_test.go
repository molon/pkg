package putil

import (
	"testing"

	"log"
)

func TestMaskMiddle(t *testing.T) {
	log.Printf("%v", MaskMiddle(""))
	log.Printf("%v", MaskMiddle("1"))
	log.Printf("%v", MaskMiddle("12"))
	log.Printf("%v", MaskMiddle("123"))
	log.Printf("%v", MaskMiddle("1234"))
	log.Printf("%v", MaskMiddle("12345"))
	log.Printf("%v", MaskMiddle("123456"))
	log.Printf("%v", MaskMiddle("1234567"))
	log.Printf("%v", MaskMiddle("12345678"))
	log.Printf("%v", MaskMiddle("6d452f938304ea12452ca403ad0b9fc4"))
}
