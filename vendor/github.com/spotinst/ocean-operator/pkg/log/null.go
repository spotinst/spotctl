// Copyright 2021 NetApp, Inc. All Rights Reserved.

package log

import (
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NullLogger is a logr.Logger that does nothing.
var NullLogger = log.NullLogger{}
