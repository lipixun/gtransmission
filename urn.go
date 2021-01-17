// Author: lipixun
// Created Time : 2021-01-17 22:09:23
//
// File Name: urn.go
// Description:
//
//	Reference:
//
//		https://en.wikipedia.org/wiki/Uniform_Resource_Name
//

package transmission

import (
	"errors"
	"fmt"
	"strings"
)

// Errors
var (
	ErrMalformedUrn = errors.New("Malformed urn")
)

// Urn defines urn
type Urn struct {
	Nid string // Namespace identifier
	Nss string // Interpretation of this field depends on the specified namespace
}

func (u Urn) String() string {
	return fmt.Sprintf("urn:%v:%v", u.Nid, u.Nss)
}

// ParseUrn parses urn
func ParseUrn(urn string) (u Urn, err error) {
	strs := strings.Split(urn, ":")
	if len(strs) != 3 {
		err = fmt.Errorf("%w", ErrMalformedUrn)
		return
	}
	var (
		scheme = strs[0]
		nid    = strs[1]
		nss    = strs[2]
	)
	if strings.ToLower(scheme) != "urn" {
		err = fmt.Errorf("%w: Invalid scheme", ErrMalformedUrn)
		return
	}
	u.Nid = nid
	u.Nss = nss
	return
}
