// Author: lipixun
// Created Time : 2021-01-17 23:52:24
//
// File Name: hash.go
// Description:
//

package transmission

// Hash type
const (
	HashSHA1   = "sha1"
	HashSHA256 = "sha256"
)

// HashValue defines the hash value
type HashValue struct {
	Type  string
	Value []byte
}
