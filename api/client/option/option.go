package option

type Option struct {
	Scenario Scenario
	Host     string
}

type Scenario struct {
	Executor        string // can be constant-arrival-rate or other
	Rate            int    // number of requests per second
	Timeunit        int    // in seconds (per second)
	Duration        int    // test duration in second
	PreAllocatedVUs int    // TODO initial VUs
	MaxVUs          int    // TODO maximum VUs
}

type Method string

// Methods
const (
	Publish   Method = "publish"
	Fetch     Method = "fetch"
	Subscribe Method = "subscribe"
)
