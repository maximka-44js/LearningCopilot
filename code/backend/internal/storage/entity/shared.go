package entity

import (
	"regexp"
)

var (
	reSafeTitle       = regexp.MustCompile(`^[\p{L}\p{N}\s\-_.,!?()«»"']{1,100}$`)
	reSafeDescription = regexp.MustCompile(`^[\p{L}\p{N}\s\-_.,!?()«»"']{0,1000}$`)
)
