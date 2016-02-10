<center>
# tokenbucket
</center>


## OVERVIEW

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


### Installation

If you have a working go installation on a Unix-like OS:

> ```go get github.com/hotei/tokenbucket```

This will copy github.com/hotei/program to the <font color=red>_first_</font> entry of your $GOPATH

or if go is not installed yet :

> ```cd DestinationDirectory```

> ```git clone https://github.com/hotei/tokenbucket.git```

### Features
* Conceptually simple
* Easy to use API
* Well [documented][4] 
* Easy to adapt (less than 100 LOC)

### Limitations

* <font color="red">TBD</font>

### Usage
```
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
	
```

The included example program provides a more interesting situation with multiple 
goroutines and examines the fairness of the tokenbucket process.

<!-- ### BUGS -->

### To-Do

* Essential:
 * TBD
* Nice:
 * TBD

### Change Log
* 2014-03-xx Started

### Resources

* [go language reference] [1] 
* [go standard library package docs] [2]
* [Source for program] [3]
* [related wikipedia article] [4]
* [godoc.org API listing][5]

[1]: http://golang.org/ref/spec/ "go reference spec"
[2]: http://golang.org/pkg/ "go package docs"
[3]: http://github.com/hotei/tokenbucket "github.com/hotei/tokenbucket"
[4]: http://en.wikipedia.org/wiki/Token_bucket
[5]: http://godoc.org/github.com/hotei/tokenbucket "godoc.org API"

Comments can be sent to <hotei1352@gmail.com> or to user "hotei" at github.com.
License is BSD-two-clause, in file "LICENSE" as well as here.

License
-------
The 'tokenbucket' go package/programs are distributed under the Simplified BSD License:

> Copyright (c) 2014 David Rook. All rights reserved.
> 
> Redistribution and use in source and binary forms, with or without modification, are
> permitted provided that the following conditions are met:
> 
>    1. Redistributions of source code must retain the above copyright notice, this list of
>       conditions and the following disclaimer.
> 
>    2. Redistributions in binary form must reproduce the above copyright notice, this list
>       of conditions and the following disclaimer in the documentation and/or other materials
>       provided with the distribution.
> 
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDER ``AS IS'' AND ANY EXPRESS OR IMPLIED
> WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
> FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> OR
> CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
> CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
> SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
> ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
> NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
> ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.



// EOF README-tokenbucket.md     (c) 2014 David Rook 
