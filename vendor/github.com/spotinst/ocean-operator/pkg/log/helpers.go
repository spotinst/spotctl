// Copyright 2021 NetApp, Inc. All Rights Reserved.

package log

import (
	"io"
)

// MaybeSetWriter will call log.SetWriter(w) if logger has a SetWriter method.
func MaybeSetWriter(log Logger, w io.Writer) {
	type writerSetter interface {
		SetWriter(io.Writer)
	}
	v, ok := log.(writerSetter)
	if ok {
		v.SetWriter(w)
	}
}

// MaybeSetVerbosity will call log.SetVerbosity(verbosity) if logger has
// a SetVerbosity method.
func MaybeSetVerbosity(log Logger, verbosity Level) {
	type verbositySetter interface {
		SetVerbosity(Level)
	}
	v, ok := log.(verbositySetter)
	if ok {
		v.SetVerbosity(verbosity)
	}
}
