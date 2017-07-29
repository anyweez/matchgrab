package structs

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func closeEnough(first float64, second float64) bool {
	return math.Abs(first-second) < 0.5
}

func TestLowRpsSerial(t *testing.T) {
	p := NewPacer(30, 1)

	count := 0
	runs := 5

	start := time.Now()
	p.Run(func() {
		count++
	}, runs)

	duration := time.Now().Sub(start).Seconds()

	if count != runs || !closeEnough(duration, 8.0) {
		t.Fail()
	}
}

func TestHighRpsSerial(t *testing.T) {
	p := NewPacer(600, 1)

	count := 0
	runs := 100

	start := time.Now()
	p.Run(func() {
		count++
	}, runs)

	duration := time.Now().Sub(start).Seconds()

	if count != runs || !closeEnough(duration, 10.0) {
		t.Fail()
	}
}

// PauseFor should stop execution of paused tasks for two seconds
func TestLowRpsSerialPaused(t *testing.T) {
	p := NewPacer(60, 1)

	count := 0
	runs := 5

	start := time.Now()
	// first run (4 seconds)
	p.Run(func() {
		count++
	}, runs)

	p.PauseFor(2 * time.Second) // pause (2 seconds)

	// second run (4 seconds)
	p.Run(func() {
		count++
	}, runs)

	duration := time.Now().Sub(start).Seconds()

	if count != runs*2 || !closeEnough(duration, 10.0) {
		fmt.Println(duration)
		t.Fail()
	}
}

func TestInfiniteRuns(t *testing.T) {
	p := NewPacer(600, 1)
	start := time.Now()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		count := 0
		p.Run(func() {
			count++

			// Manually close after 100 executions
			if count > 100 {
				wg.Done()
			}
		}, 0)
	}()

	// Test fails if we've been running for 15 seconds.
	go func() {
		for {
			if time.Now().Sub(start).Seconds() > 15.0 {
				t.Fail()
			}

			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()

	duration := time.Now().Sub(start).Seconds()

	if !closeEnough(duration, 10.0) {
		t.Fail()
	}
}
