// Author: lipixun
// Created Time : 2021-01-17 22:44:13
//
// File Name: num_range.go
// Description:
//

package transmission

import (
	"errors"
	"strconv"
	"strings"
)

// NumRange defines a number range
type NumRange struct {
	Start        int
	End          int
	IncludeStart bool
	IncludeEnd   bool
}

// NewSingleNumRange creates a new NumRange by a single number
func NewSingleNumRange(num int) NumRange {
	return NumRange{num, num, true, true}
}

// ParseNumRangeFromString parses a number range from string
// Format:
//	\d+
//	\d+\-\d+
func ParseNumRangeFromString(s string) (r NumRange, err error) {
	strs := strings.Split(s, "-")
	if len(strs) == 1 {
		var num int
		num, err = strconv.Atoi(s)
		if err != nil {
			return
		}
		r.Start = num
		r.End = num
		r.IncludeStart = true
		r.IncludeEnd = true
		return
	} else if len(strs) == 2 {
		var num int
		// Start
		num, err = strconv.Atoi(strs[0])
		if err != nil {
			return
		}
		r.Start = num
		r.IncludeStart = true
		// End
		num, err = strconv.Atoi(strs[1])
		if err != nil {
			return
		}
		r.End = num
		r.IncludeEnd = true
	}

	err = errors.New("Malformed num range string")
	return
}
