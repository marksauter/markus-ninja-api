package oid_test

import (
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

var testOID, _ = oid.New("Test")

func TestNewFromShort(t *testing.T) {
	id, err := oid.NewFromShort("Test", testOID.Short)
	if err != nil {
		t.Errorf(
			"TestNewFromShort(%s): unexpected err: %s",
			testOID.Short,
			err,
		)
	}
	expected := testOID.String
	actual := id.String
	if actual != expected {
		t.Errorf(
			"TestNewFromShort(%s): expected %s, actual %s",
			testOID.Short,
			expected,
			actual,
		)
	}
}

func TestParse(t *testing.T) {
	id, err := oid.Parse(testOID.String)
	if err != nil {
		t.Errorf(
			"TestParse(%s): unexpected err: %s",
			testOID.String,
			err,
		)
	}
	expected := testOID
	actual := id
	if *actual != *expected {
		t.Errorf(
			"TestParse(%s): expected %+v, actual %+v",
			testOID.String,
			expected,
			actual,
		)
	}
}
