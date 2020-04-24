package diff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"
)

type equalOpts struct {
	timeTolerance time.Duration
	headerFields  []string
}

//EqualOption is an option for Equal
type EqualOption func(*equalOpts)

//TimeTolerance sets some tolerance so timing doesn't have to be exactly equal between two files
func TimeTolerance(tolerance time.Duration) EqualOption {
	return func(opts *equalOpts) {
		opts.timeTolerance = tolerance
	}
}

//CompareHeaderFields sets a list of headers to compare
func CompareHeaderFields(fields ...string) EqualOption {
	return func(opts *equalOpts) {
		opts.headerFields = append(opts.headerFields, fields...)
	}
}

//Equal tests whether two asciinema cast files are equal
func Equal(a, b io.Reader, opt ...EqualOption) (bool, error) {
	opts := new(equalOpts)
	for _, o := range opt {
		o(opts)
	}
	aScanner := bufio.NewScanner(a)
	bScanner := bufio.NewScanner(b)
	var lineCount int
	var lastATime, lastBTime time.Duration
	for aScanner.Scan() {
		var err error
		if !bScanner.Scan() {
			return false, nil
		}
		lineCount++
		if lineCount == 1 {
			var aHeader, bHeader map[string]interface{}
			err = json.Unmarshal(aScanner.Bytes(), &aHeader)
			if err != nil {
				return false, err
			}
			err = json.Unmarshal(bScanner.Bytes(), &bHeader)
			if err != nil {
				return false, err
			}
			if !compareHeaders(aHeader, bHeader, opts.headerFields...) {
				return false, nil
			}
			continue
		}
		aEvent := new(event)
		bEvent := new(event)
		err = json.Unmarshal(aScanner.Bytes(), aEvent)
		if err != nil {
			return false, err
		}
		err = json.Unmarshal(bScanner.Bytes(), bEvent)
		if err != nil {
			return false, err
		}
		aTimeDiff := aEvent.Time - lastATime
		wantBTime := lastBTime + aTimeDiff

		wantB := &event{
			Time: wantBTime,
			Type: aEvent.Type,
			Data: aEvent.Data,
		}

		if !wantB.equal(bEvent, opts.timeTolerance) {
			return false, nil
		}

		lastATime = aEvent.Time
		lastBTime = bEvent.Time
	}
	if bScanner.Scan() {
		return false, nil
	}
	return true, nil
}

func compareHeaders(a, b map[string]interface{}, fields ...string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		a = map[string]interface{}{}
	}
	if b == nil {
		b = map[string]interface{}{}
	}
	for _, field := range fields {
		if !reflect.DeepEqual(a[field], b[field]) {
			return false
		}
	}
	return true
}

type event struct {
	Time time.Duration
	Type string
	Data string
}

func (e *event) equal(other *event, timeTolerance time.Duration) bool {
	if e == nil || other == nil {
		return e == nil && other == nil
	}
	if e.Type != other.Type || e.Data != other.Data {
		return false
	}
	if other.Time < e.Time-timeTolerance {
		return false
	}
	if other.Time > e.Time+timeTolerance {
		return false
	}
	return true
}

func (e *event) UnmarshalJSON(data []byte) error {
	jsonSlice := make([]interface{}, 0, 3)
	err := json.Unmarshal(data, &jsonSlice)
	if err != nil {
		return err
	}
	if len(jsonSlice) != 3 {
		return fmt.Errorf("invalid event data")
	}
	seconds, ok := jsonSlice[0].(float64)
	if !ok {
		return fmt.Errorf("invalid time")
	}
	e.Time = time.Duration(seconds * float64(time.Second))
	e.Type, ok = jsonSlice[1].(string)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	e.Data, ok = jsonSlice[2].(string)
	if !ok {
		return fmt.Errorf("invalid data")
	}
	return nil
}
