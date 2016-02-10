/*
The tokenbucket package is a rate limiter based on the concept of a bucket being
filled with tokens at a steady rate.

Once the bucket is filled to capacity, no more tokens can be added.

Tokens are withdrawn by cooperating functions that call Take(nTokens) which
returns the amount of time the caller should sleep before sufficient tokens
will be available.

Once withdrawn/reserved tokens can not be put back, thus negative token counts
are not allowed.

A caller can request more tokens than the capacity of the bucket but the delay
returned will be proportionally longer.

Initialization requires a fill rate expressed as the delay between increments,
and a bucket capacity.

The rate per second is 1 / delay-in-seconds.  A convenience function for this is
FillRate() float64

Nothing forces the participants to cooperate.  This could be a bug or a feature
depending on how you intend to use it.  Goroutines generated inside a for loop
can all share the same bucket easily as the example.go program demonstrates.

// bucket that fills at rate of 10 units per second, max of 200 units
	tb := tokenbucket.New(time.Millisecond * 100, 200)

// tokens are often used to rate-limit tight loops
// this example will allow a short burst (using up the 200) before it throttles down
// to the 10 unit per second limit
	for {
		time.Sleep(tb.Take(5))
		... do something ...
	}

// similar but the capacity is used up immediately so even the first one will wait
	for {
		time.Sleep(tb.Take(500))
		... do something ...
	}

// start with empty bucket
	tb := tokenbucket.New(time.Millisecond * 100, 200)
	tb.Take(200)

*/
package tokenbucket
