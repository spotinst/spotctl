// Copyright 2021 NetApp, Inc. All Rights Reserved.

package log

import "github.com/go-logr/logr"

// Level is a verbosity logging level for Info logs.
// See also https://github.com/kubernetes/klog
type Level int32

// Logger describes the interface that must be implemented by all loggers.
type Logger = logr.Logger
