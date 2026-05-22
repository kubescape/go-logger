package otelsetup

import (
	"context"
	"sync"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// RingBufferLogProcessor keeps the last 7500 log records in memory so
// operators can flush them retroactively via the debug HTTP listener after a
// suspicious event is observed.
type RingBufferLogProcessor struct {
	buf      [7500]sdklog.Record
	head     int
	tail     int
	size     int
	flushing bool // true while FlushToBackend is re-emitting; drops incoming records to break the loop
	mu       sync.Mutex
}

// OnEmit clones the record (sdk/log Records are not concurrent-safe — the
// upstream BatchProcessor may mutate them after our return) and inserts it
// into the ring buffer. Records are dropped silently while a flush is in
// progress to prevent re-emitted records from looping back into the buffer.
func (p *RingBufferLogProcessor) OnEmit(_ context.Context, r *sdklog.Record) error {
	clone := r.Clone()
	p.mu.Lock()
	if p.flushing {
		p.mu.Unlock()
		return nil
	}
	p.buf[p.head] = clone
	p.head = (p.head + 1) % len(p.buf)
	if p.size < len(p.buf) {
		p.size++
	} else {
		p.tail = (p.tail + 1) % len(p.buf)
	}
	p.mu.Unlock()
	return nil
}

// Enabled returns true for Info-level and above. Debug records are excluded
// to keep the ~1.5 MB memory bound realistic and to avoid storing high-volume
// hot-path debug output.
// NOTE: full package-level filtering (allowlist containerprofilemanager,
// sbommanager, objectcache, exporters; exclude rulemanager, containerwatcher)
// requires implementing the ring buffer as a slog.Handler so Enabled fires
// before the slog record is constructed. The severity gate here is the
// best approximation available at the OTEL processor layer.
func (p *RingBufferLogProcessor) Enabled(_ context.Context, params sdklog.EnabledParameters) bool {
	return params.Severity >= otellog.SeverityInfo1
}

// Shutdown is a no-op — the buffer is in-memory only.
func (p *RingBufferLogProcessor) Shutdown(_ context.Context) error { return nil }

// ForceFlush is a no-op for the same reason.
func (p *RingBufferLogProcessor) ForceFlush(_ context.Context) error { return nil }

// FlushToBackend drains the ring buffer and re-emits the captured records
// through the provided log.Logger so the LoggerProvider's existing
// BatchProcessor handles delivery. The buffer is cleared atomically before
// re-emitting, and incoming records are gated during the flush so re-emitted
// records do not loop back into the buffer — a second call exports only
// records that arrived after the first flush completed.
func (p *RingBufferLogProcessor) FlushToBackend(ctx context.Context, l otellog.Logger) {
	p.mu.Lock()
	records := make([]sdklog.Record, p.size)
	for i := range p.size {
		records[i] = p.buf[(p.tail+i)%len(p.buf)]
	}
	p.head = 0
	p.tail = 0
	p.size = 0
	p.flushing = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.flushing = false
		p.mu.Unlock()
	}()

	for i := range records {
		l.Emit(ctx, sdkRecordToLogRecord(&records[i]))
	}
}

func sdkRecordToLogRecord(r *sdklog.Record) otellog.Record {
	var out otellog.Record
	out.SetTimestamp(r.Timestamp())
	out.SetObservedTimestamp(r.ObservedTimestamp())
	out.SetSeverity(r.Severity())
	out.SetSeverityText(r.SeverityText())
	out.SetBody(r.Body())
	out.SetEventName(r.EventName())
	r.WalkAttributes(func(kv otellog.KeyValue) bool {
		out.AddAttributes(kv)
		return true
	})
	return out
}
