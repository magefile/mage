package mg

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type stream int

const (
	stdoutStream stream = 1
	stderrStream stream = 2
)

type outputLine struct {
	stream    stream
	text      string
	timeStart time.Time
	timeEnd   time.Time
}

// outputCollector collects whole lines of output
type outputCollector interface {
	AddLine(line outputLine)
}

// streamLineSink constructs whole lines of data for a single output stream and
// adds them to the collector
type streamLineSink struct {
	collector outputCollector
	stream    stream

	// current non-\n-terminated line of output
	tailStartTime time.Time
	tail          string
}

func (sls *streamLineSink) Add(t time.Time, input string) {
	for _, line := range strings.SplitAfter(input, "\n") {
		if strings.HasSuffix(line, "\n") {
			// full line
			var ol outputLine
			if sls.tail == "" {
				ol = outputLine{
					stream:    sls.stream,
					text:      line,
					timeStart: t,
					timeEnd:   t,
				}
			} else {
				ol = outputLine{
					stream:    sls.stream,
					text:      sls.tail + line,
					timeStart: sls.tailStartTime,
					timeEnd:   t,
				}
				sls.tail = ""
			}
			sls.collector.AddLine(ol)
		} else {
			// partial line, may only happen once at the end of the loop
			if sls.tail == "" {
				sls.tail = line
				sls.tailStartTime = t
			} else {
				sls.tail += line
			}
		}
	}
}

func (sls *streamLineSink) Flush(t time.Time) {
	if sls.tail != "" {
		ol := outputLine{
			stream:    sls.stream,
			text:      sls.tail + "\n",
			timeStart: sls.tailStartTime,
			timeEnd:   t,
		}
		sls.collector.AddLine(ol)
		sls.tail = ""
	}
}

// streamLineWriter is a Writer for an output stream that flushes any pending
// data from other stream
type streamLineWriter struct {
	sink  *streamLineSink
	other *streamLineSink
}

func (slw streamLineWriter) Write(p []byte) (int, error) {
	t := time.Now()
	slw.other.Flush(t)
	slw.sink.Add(t, string(p))
	return len(p), nil
}

func (slw streamLineWriter) Flush() {
	slw.sink.Flush(time.Now())
}

type flushWriter interface {
	io.Writer
	Flush()
}

func newStreamLineWriters(collector outputCollector) (flushWriter, flushWriter) {
	stdoutSink := &streamLineSink{collector: collector, stream: stdoutStream}
	stderrSink := &streamLineSink{collector: collector, stream: stderrStream}

	return streamLineWriter{stdoutSink, stderrSink}, streamLineWriter{stderrSink, stdoutSink}
}

type taskOutputCollector struct {
	taskID   int
	taskName string
}

var streamTag = map[stream]string{
	stdoutStream: " ",
	stderrStream: "E",
}

const slowLineThreshold = 100 * time.Millisecond
const timeFormat = "2006-01-02 15:04:05.000Z07:00"

func (toc taskOutputCollector) AddLine(line outputLine) {
	dur := line.timeEnd.Sub(line.timeStart)
	tags := ""
	if dur >= slowLineThreshold {
		tags = fmt.Sprintf(" slow-output-line=%dms", dur/time.Millisecond)
	}

	fmt.Printf("%s %s #%04d %s%s | %s", line.timeStart.Format(timeFormat), streamTag[line.stream], toc.taskID, toc.taskName, tags, line.text)
}
