package tracer

import "time"

type traceBatchData struct {
	batchStartTime time.Time
	queryStartTime time.Time
}

func (d *traceBatchData) setQueryStartTime() {
	d.queryStartTime = time.Now()
}
