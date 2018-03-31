package main

// Simple tests for the bannergrabtroll.go file

import (
	"reflect"
	"testing"
)

// This method runs unit16Ranges against the input and compares it to the expected
// This assumes no error should be thrown. For tests that should throw an error,
// use assertionErrTest
func assertionNoErrTest(input string, expected [][]uint16, t *testing.T) {
	actual, err := unit16Ranges(input)
	if err != nil {
		t.Fatalf("Unexpected error %s", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected %v but got %v", expected, actual)
	}
}

// Test that assumes there is an error in the input string to unit16Ranges
// (and fails if there is no error)
func assertionErrTest(input string, t *testing.T) {
	_, err := unit16Ranges(input)
	if err == nil {
		t.Fatalf("Did not get an error from input %s", input)
	}
}

// Test a single invalid value
func TestErrInvalidSingle(t *testing.T) {
	assertionErrTest("x", t)
}

// Test a single out of range value on the high side
func TestErrInvalidSingleOutOfRangeHigh(t *testing.T) {
	assertionErrTest("65536", t)
}

// Test empty string
func TestEmptyString(t *testing.T) {
	assertionErrTest("", t)
}

// Test multiple values where an invalid is inside one of them
func TestMultipleOneInvalid(t *testing.T) {
	assertionErrTest("1,x,2", t)
}

// Test a completely invalid minimum for a range
func TestInvalidMinimum(t *testing.T) {
	assertionErrTest("x-10", t)
}

// Test a completely invalid maximum for a range
func TestInvalidMaximum(t *testing.T) {
	assertionErrTest("9-x", t)
}

// Test a single out of range value on the low side
func TestErrInvalidSingleOutOfRangeLow(t *testing.T) {
	assertionErrTest("-1", t)
}

// Test for one valid value at the minimum
func TestUnit16RangesValidOneMin(t *testing.T) {
	assertionNoErrTest("0", [][]uint16{{0}}, t)
}

// Test for one valid value at the maximum
func TestUnit16RangesValidOneMax(t *testing.T) {
	assertionNoErrTest("65535", [][]uint16{{65535}}, t)
}

// Test for one range covering all values
func TestUnit16RangesValidAllValues(t *testing.T) {
	assertionNoErrTest("0-65535", [][]uint16{{0, 65535}}, t)
}

// Test for two disjoint ranges
func TestUnit16RangesDisjointRanges(t *testing.T) {
	assertionNoErrTest("1-5,10-12", [][]uint16{{1, 5}, {10, 12}}, t)
}

// Test for two single values
func TestUnit16RangesTwoValues(t *testing.T) {
	assertionNoErrTest("22,80", [][]uint16{{22}, {80}}, t)
}
