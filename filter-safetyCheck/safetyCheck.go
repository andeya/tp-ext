// Package safetyCheck provides a integrity check transfer filter
package safetyCheck

import (
	"bytes"
	"crypto/md5"
	"errors"
)

// safetyCheck compression filter
type safetyCheck struct{}

const md5Length = 16

var errDataCheck = errors.New("check failed")

// NewSafetyCheck returns a safetyCheck object
func NewSafetyCheck() *safetyCheck {
	return &safetyCheck{}
}

// Id returns transfer filter id.
func (s *safetyCheck) Id() byte {
	return 0
}

func (s *safetyCheck) OnPack(src []byte) ([]byte, error) {
	content, err := getMd5(src)
	if err != nil {
		return nil, err
	}
	src = append(src, content...)

	return src, nil
}

func (s *safetyCheck) OnUnpack(src []byte) ([]byte, error) {
	srcLength := len(src)
	if srcLength < md5Length {
		return nil, errDataCheck
	}
	srcData := src[:srcLength-md5Length]
	content, err := getMd5(srcData)
	if err != nil {
		return nil, err
	}
	// Check
	if !bytes.Equal(content, src[srcLength-md5Length:]) {
		return nil, errDataCheck
	}
	return srcData, nil
}

func getMd5(src []byte) ([]byte, error) {
	md5Hash := md5.New()
	_, err := md5Hash.Write(src)
	if err != nil {
		return nil, err
	}

	return md5Hash.Sum(nil), nil
}
