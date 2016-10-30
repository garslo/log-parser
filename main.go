package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type TimedLog struct {
	Type     string
	Duration time.Duration
}

func (me TimedLog) FullyParsed() bool {
	return me.Type != "" && me.Duration != 0
}

func (me TimedLog) String() string {
	return fmt.Sprintf(`type="%s" duration_ms=%f`, me.Type, me.Duration.Seconds()*1000)
}

type SingleFileSource struct {
	Filename           string
	MaybeDurationField func([]byte) (time.Duration, bool)
	MaybeTypeField     func([]byte) (string, bool)
}

func (me *SingleFileSource) Iter() (chan TimedLog, error) {
	fh, err := os.Open(me.Filename)
	if err != nil {
		return nil, err
	}
	iter := make(chan TimedLog)
	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)
	go func() {
		defer fh.Close()
		for scanner.Scan() {
			line := scanner.Bytes()
			fields := bytes.Fields(line)
			logLine := TimedLog{}
			for _, field := range fields {
				if logLine.FullyParsed() {
					break
				}
				if duration, ok := me.MaybeDurationField(field); ok {
					logLine.Duration = duration
					continue
				}
				if type_, ok := me.MaybeTypeField(field); ok {
					logLine.Type = type_
					continue
				}
			}
			if !logLine.FullyParsed() {
				continue
			}
			iter <- logLine
		}
		close(iter)
	}()
	return iter, nil
}

func MakeMaybeTypeField(keyName []byte) func([]byte) (string, bool) {
	return func(in []byte) (string, bool) {
		if !bytes.HasPrefix(in, keyName) {
			return "", false
		}
		split := bytes.Split(in, []byte("="))
		if len(split) != 2 {
			return "", false
		}
		return string(split[1]), true
	}
}

var (
	DurationMSPrefix = []byte("duration_ms=")
	ExecMSPrefix     = []byte("exec_ms=")
	ExecPrefix       = []byte("exec=")
	DurationPrefix   = []byte("duration=")
	TimeMSPrefix     = []byte("time_ms=")
	TimePrefix       = []byte("time=")
)

func MaybeDurationField(in []byte) (time.Duration, bool) {
	if bytes.HasPrefix(in, DurationMSPrefix) {
		return ParseMS(bytes.Split(in, []byte("="))[1])
	}
	if bytes.HasPrefix(in, ExecMSPrefix) {
		return ParseMS(bytes.Split(in, []byte("="))[1])
	}
	if bytes.HasPrefix(in, TimeMSPrefix) {
		return ParseMS(bytes.Split(in, []byte("="))[1])
	}
	if bytes.HasPrefix(in, DurationPrefix) {
		return ParseDuration(bytes.Split(in, []byte("="))[1])
	}
	if bytes.HasPrefix(in, ExecPrefix) {
		return ParseDuration(bytes.Split(in, []byte("="))[1])
	}
	if bytes.HasPrefix(in, TimeMSPrefix) {
		return ParseDuration(bytes.Split(in, []byte("="))[1])
	}
	return time.Millisecond, false
}

func ParseMS(ms []byte) (time.Duration, bool) {
	msFloat, err := strconv.ParseFloat(string(ms), 64)
	if err != nil {
		return time.Millisecond, false
	}
	return time.Duration(int64(msFloat)) * time.Millisecond, true
}

func ParseDuration(ms []byte) (time.Duration, bool) {
	dur, err := time.ParseDuration(string(ms))
	return dur, err == nil
}

func main() {
	var (
		key   string
		value string
	)

	flag.StringVar(&key, "key", "mode", "key for identifying log type")
	flag.StringVar(&value, "value", "", "value for identifying log type")
	flag.Parse()

	writer := make(chan float64)
	wg := sync.WaitGroup{}
	for _, filename := range flag.Args() {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()
			source := &SingleFileSource{fname, MaybeDurationField, MakeMaybeTypeField([]byte(key))}
			iter, err := source.Iter()
			if err != nil {
				log.Print(err.Error())
				return
			}
			for logLine := range iter {
				writer <- logLine.Duration.Seconds() * 1000
			}
		}(filename)
	}

	go func() {
		for toWrite := range writer {
			fmt.Printf("%.2f\n", toWrite)
		}
	}()
	wg.Wait()
}
