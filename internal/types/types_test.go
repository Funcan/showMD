package types_test

import (
	"testing"

	"github.com/funcan/showmd/internal/types"
)

func TestBool_NilReturnsDefault(t *testing.T) {
	if got := types.Bool(nil, true); got != true {
		t.Errorf("Bool(nil, true): want true, got %v", got)
	}
	if got := types.Bool(nil, false); got != false {
		t.Errorf("Bool(nil, false): want false, got %v", got)
	}
}

func TestBool_NonNilReturnsValue(t *testing.T) {
	tr, fl := true, false
	if got := types.Bool(&tr, false); got != true {
		t.Errorf("Bool(&true, false): want true, got %v", got)
	}
	if got := types.Bool(&fl, true); got != false {
		t.Errorf("Bool(&false, true): want false, got %v", got)
	}
}

func TestInt_NilReturnsDefault(t *testing.T) {
	if got := types.Int(nil, 42); got != 42 {
		t.Errorf("Int(nil, 42): want 42, got %v", got)
	}
}

func TestInt_NonNilReturnsValue(t *testing.T) {
	v := 7
	if got := types.Int(&v, 99); got != 7 {
		t.Errorf("Int(&7, 99): want 7, got %v", got)
	}
}

func TestInt_ZeroValue(t *testing.T) {
	v := 0
	if got := types.Int(&v, 10); got != 0 {
		t.Errorf("Int(&0, 10): want 0 (not default), got %v", got)
	}
}

func TestString_NilReturnsDefault(t *testing.T) {
	if got := types.String(nil, "default"); got != "default" {
		t.Errorf("String(nil, \"default\"): want \"default\", got %q", got)
	}
}

func TestString_NonNilReturnsValue(t *testing.T) {
	v := "hello"
	if got := types.String(&v, "default"); got != "hello" {
		t.Errorf("String(&\"hello\", \"default\"): want \"hello\", got %q", got)
	}
}

func TestString_EmptyStringNotDefault(t *testing.T) {
	v := ""
	if got := types.String(&v, "default"); got != "" {
		t.Errorf("String(&\"\", \"default\"): want \"\" (not default), got %q", got)
	}
}
