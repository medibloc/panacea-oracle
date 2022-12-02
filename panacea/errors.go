package panacea

import "fmt"

var (
	ErrEmptyKey             = fmt.Errorf("empty key")
	ErrEmptyValue           = fmt.Errorf("empty value")
	ErrNegativeOrZeroHeight = fmt.Errorf("negative or zero height")
)
