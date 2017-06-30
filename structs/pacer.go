package structs

import "time"

// Pacer : Runs a function at a specified rate. In matchgrab this is being used for making
// API requests to Riot but I think its written generically enough that it could be repurposed
// for something else as well.
//
// When a new pacer is created, a goroutine pool is also launched that monitor the input queue
// of functions to be executed (added w/ Each() function call). When executing functions from the
// queue, it will execute up to `maxSimultaneousRequests` functions simultaneously; if you want
// to avoid this simply set the value to 1 at initialization.
//
// You can also pause execution for any period using the PauseFor() function.
type Pacer struct {
	// config options
	requestsPerMinute       int
	maxSimultaneousRequests int

	sim       chan bool // keep track of how many are running now
	next      chan bool
	pausedFor time.Duration

	queue chan func()
}

func NewPacer(rpm int, sim int) *Pacer {
	p := &Pacer{
		requestsPerMinute:       rpm,
		maxSimultaneousRequests: sim,

		// sim:  make(chan bool, sim),
		next:      make(chan bool),
		queue:     make(chan func()),
		pausedFor: 0, // not required, but explicit > implicit :)
	}

	// Start pacing goroutine
	go func() {
		delay := int((60.0 / float32(p.requestsPerMinute)) * 1000)

		for {
			// Don't block; if the channel isn't empty, skip and wait.
			if len(p.next) == 0 {
				p.next <- true
			}

			// Don't call any new functions if a pause has been specified.
			if p.pausedFor != 0 {
				time.Sleep(p.pausedFor)
				p.pausedFor = 0
			} else {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}()

	// Start up the goroutine pool
	for i := 0; i < sim; i++ {
		go func() {
			// Wait for a task to do. Once we get it, block until allowed to proceed
			// based on pacing rules.
			for fn := range p.queue {

				<-p.next
				fn()
			}
		}()
	}

	return p
}

// PauseFor : Pauses the pacer and will not start any new executions until the specified duration
// passes.
func (p *Pacer) PauseFor(d time.Duration) {
	p.pausedFor = d
}

// Each : Runs the specific function as quickly as allowed w/ pacing rules. A pacer starts
// each run on a separate goroutine (up to maxSimultaneousRequests at a time) so its
// likely that multiple instances will be running at once if that's > 1.
//
// If count is zero, runs indefinitely.
func (p *Pacer) Run(fn func(), count int) {
	if count == 0 { // infinite
		for {
			p.queue <- fn
		}
	} else { // finite # of calls
		for i := 0; i < count; i++ {
			p.queue <- fn
		}
	}
}
