// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"time"

	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-stack/stack"
)

const (
	// To get the Caller information correct on the log, we need to count the
	// number of calls from a log call in the code to the time it hits a kitlog
	// context: [log call site (5), Info/Trace (4), MultipleChannelLogger.Log (3),
	// kitlog.Context.Log (2), kitlog.bindValues (1) (binding occurs),
	// kitlog.Caller (0), stack.caller]
	infoTraceLoggerCallDepth = 5
)

var defaultTimestampUTCValuer kitlog.Valuer = func() interface{} {
	return time.Now()
}

func WithMetadata(infoTraceLogger loggers.InfoTraceLogger) loggers.InfoTraceLogger {
	return infoTraceLogger.With(structure.TimeKey, defaultTimestampUTCValuer,
		structure.CallerKey, kitlog.Caller(infoTraceLoggerCallDepth),
		"trace", TraceValuer())
}

func TraceValuer() kitlog.Valuer {
	return func() interface{} { return stack.Trace() }
}
