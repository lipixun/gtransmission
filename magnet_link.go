// Author: lipixun
// Created Time : 2021-01-17 21:03:47
//
// File Name: magnet.go
// Description:
//
//	Reference:
//
//		https://en.wikipedia.org/wiki/Magnet_URI_scheme
//		https://www.bittorrent.org/beps/bep_0053.html
//

package transmission

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

//
//
//
// Magnet link
//
//
//

// Errors
var (
	ErrMalformedMagnetLink = errors.New("Malformed magnet link")
	ErrWrongMagnetLinkType = errors.New("Wrong magnet link type")
)

// MagnetLink defines magnet link
type MagnetLink struct {
	Dn       []string            // Display name
	Xt       []Urn               // Exact topic
	Xl       []int               // Exact length
	As       []string            // Acceptable source
	Xs       []string            // Exact source
	Kt       []string            // Keyword topic
	Mt       []string            // Manifest topic
	Tr       []string            // Tracker address
	So       []NumRange          // Select only
	Exps     map[string][]string // Experimental parameters (which must begin with "x.")
	Unknowns map[string][]string // Uknown parameters
}

// ParseMagnetLink parses magnetLink uri
func ParseMagnetLink(uri string, opts ...MagnetLinkParseOption) (*MagnetLink, error) {
	var option magnetLinkParseOption
	for _, opt := range opts {
		if opt != nil {
			opt.set(&option)
		}
	}

	// Parse uri
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformedMagnetLink, err)
	}
	if strings.ToLower(u.Scheme) != "magnet" {
		return nil, fmt.Errorf("%w: Invalid scheme", ErrMalformedMagnetLink)
	}

	// Parse parameters
	var magnetLink MagnetLink
	q := u.Query()
	for key, values := range q {
		key = strings.ToLower(key)
		if key == "dn" {
			magnetLink.Dn = append(magnetLink.Dn, values...)
		} else if checkIsMagnetLinkXTParameter(key) {
			for _, value := range values {
				urn, err := ParseUrn(value)
				if err != nil {
					return nil, fmt.Errorf("%w: Invalid xt [%v]", ErrMalformedMagnetLink, err)
				}
				magnetLink.Xt = append(magnetLink.Xt, urn)
			}
		} else if key == "xl" {
			for _, value := range values {
				num, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("%w: Invalid xl [%v]", ErrMalformedMagnetLink, err)
				}
				magnetLink.Xl = append(magnetLink.Xl, num)
			}
		} else if key == "as" {
			for _, value := range values {
				value, err := url.QueryUnescape(value)
				if err != nil {
					return nil, fmt.Errorf("%w: Invalid as [%v]", ErrMalformedMagnetLink, err)
				}
				magnetLink.As = append(magnetLink.As, value)
			}
		} else if key == "xs" {
			magnetLink.Xs = append(magnetLink.Xs, values...)
		} else if key == "kt" {
			magnetLink.Kt = append(magnetLink.Kt, values...)
		} else if key == "mt" {
			magnetLink.Mt = append(magnetLink.Mt, values...)
		} else if key == "tr" {
			for _, value := range values {
				value, err := url.QueryUnescape(value)
				if err != nil {
					return nil, fmt.Errorf("%w: Invalid tr [%v]", ErrMalformedMagnetLink, err)
				}
				magnetLink.Tr = append(magnetLink.Tr, value)
			}
		} else if key == "so" {
			for _, value := range values {
				strs := strings.Split(value, ",")
				for _, s := range strs {
					numRange, err := ParseNumRangeFromString(s)
					if err != nil {
						return nil, fmt.Errorf("%w: Invalid so [%v]", ErrMalformedMagnetLink, err)
					}
					magnetLink.So = append(magnetLink.So, numRange)
				}
			}
		} else if strings.HasPrefix(key, "x.") {
			if len(key) <= 2 {
				return nil, fmt.Errorf("%w: Invalid experimental parameter", ErrMalformedMagnetLink)
			}
			key := key[2:]
			if magnetLink.Exps == nil {
				magnetLink.Exps = make(map[string][]string)
			}
			magnetLink.Exps[key] = append(magnetLink.Exps[key], values...)
		} else {
			if option.Strict {
				return nil, fmt.Errorf("%w: Uknown parameters", ErrMalformedMagnetLink)
			}
			// Unknown parameters
			if magnetLink.Unknowns == nil {
				magnetLink.Unknowns = make(map[string][]string)
			}
			magnetLink.Unknowns[key] = append(magnetLink.Unknowns[key], values...)
		}
	}

	return &magnetLink, nil
}

func checkIsMagnetLinkXTParameter(key string) bool {
	if !strings.HasPrefix(key, "xt") {
		return false
	}
	if key == "xt" {
		return true
	}
	if len(key) < 4 {
		return false
	}
	if key[2] != '.' {
		return false
	}
	_, err := strconv.Atoi(key[3:])
	return err == nil
}

// AsTorrent converts to TorrentMagnetLink
func (l *MagnetLink) AsTorrent() (*TorrentMagnetLink, error) {
	torrentMagnetLink := TorrentMagnetLink{MagnetLink: l}
	for _, xt := range l.Xt {
		if strings.ToLower(xt.Nid) == "btih" {
			var (
				err       error
				hashValue HashValue
			)
			switch len(xt.Nss) {
			case 32:
				// SHA-1. Base32 encoding
				hashValue.Type = HashSHA1
				hashValue.Value, err = base32.StdEncoding.DecodeString(xt.Nss)
			case 40:
				// SHA-1. Hex encoding
				hashValue.Type = HashSHA1
				hashValue.Value, err = hex.DecodeString(xt.Nss)
			case 56:
				// SHA-64. Base32 encoding
				hashValue.Type = HashSHA256
				hashValue.Value, err = base32.StdEncoding.DecodeString(xt.Nss)
			case 64:
				// SHA-64. Hex encoding
				hashValue.Type = HashSHA256
				hashValue.Value, err = hex.DecodeString(xt.Nss)
			default:
				return nil, fmt.Errorf("%w: Cannot decode btih [Bad length]", ErrMalformedMagnetLink)
			}
			if err != nil {
				return nil, fmt.Errorf("%w: Cannot decode btih [%v]", ErrMalformedMagnetLink, err)
			}
			torrentMagnetLink.InfoHashs = append(torrentMagnetLink.InfoHashs, hashValue)
		}
	}

	if len(torrentMagnetLink.InfoHashs) == 0 {
		return nil, fmt.Errorf("%w: No torrent", ErrWrongMagnetLinkType)
	}
	return &torrentMagnetLink, nil
}

//
//
//
// Torrent magnet link
//
//
//

// TorrentMagnetLink defines torrent magnet link
type TorrentMagnetLink struct {
	*MagnetLink

	InfoHashs []HashValue
}

// ParseTorrentMagnetLink parses torrent magnet link
func ParseTorrentMagnetLink(uri string, opts ...MagnetLinkParseOption) (*TorrentMagnetLink, error) {
	magnetLink, err := ParseMagnetLink(uri, opts...)
	if err != nil {
		return nil, err
	}
	return magnetLink.AsTorrent()
}

//
//
//
// Options
//
//
//

// MagnetLinkParseOption defines the magnet link parse option
type MagnetLinkParseOption interface {
	set(option *magnetLinkParseOption)
}
type magnetLinkParseOption struct {
	Strict bool
}
type magnetLinkParseOptionSetterFunc func(options *magnetLinkParseOption)
type magnetLinkParseOptionSetter struct {
	f magnetLinkParseOptionSetterFunc
}

func (setter magnetLinkParseOptionSetter) set(option *magnetLinkParseOption) {
	setter.f(option)
}

// WithMagnetLinkParseStrictOption defines the strict option
func WithMagnetLinkParseStrictOption(strict bool) MagnetLinkParseOption {
	return magnetLinkParseOptionSetter{
		func(option *magnetLinkParseOption) {
			option.Strict = strict
		},
	}
}
