# Miki's Optimization Guide

## [Rules of Optimization Club](https://wiki.c2.com/?RulesOfOptimizationClub)

1. You do not optimize.
2. You do not optimize, without measuring first.
3. When the performance is not bound by the code, but by external factors, the optimization is over.
4. Only optimize code that already has full unit test coverage.
5. One factor at a time.
6. No unresolved bugs, no schedule pressure.
7. Testing will go on as long as it has to.
8. If this is your first night at Optimization Club, you have to write a test case.

See more [here](https://users.ece.utexas.edu/~adnan/pike.html)

## Performance Mantras

By [Craig Hanson and Pat Crain](http://www.brendangregg.com/blog/2018-06-30/benchmarking-checklist.html)

1. Don't do it
: Can we avoid doing the calculation at all? For example: Do we need to parse the input or just pass it as-is?

2. Do it, but don't do it again
: Can we use [memoization](https://en.wikipedia.org/wiki/Memoization)/caching?
Parse objects once at the "edges" and use the parsed objects internally.

3. Do it less
: Do we need to run this every millisecond? Can every second work? Can we use only a subset of the data?

4. Do it later
: Can we make this API call async?

5. Do it when they're not looking
: Can we run the calculation in the background while doing another task?

6. Do it concurrently
: Will concurrency help here? Consider [Amdhal's law](https://en.wikipedia.org/wiki/Amdahl%27s_law).

7. Do it cheaper
: Can we use a map here instead of a slice? Research available algorithms and data structures and know their complexity. Test them on *your* data


## Go Specific

1. Memory Allocation
: Avoid allocations as possible (see the design of [io.Reader](https://golang.org/pkg/io/#Reader)). Pre-allocate if you already know the size. Be careful of slices keep large amounts of memory (`s := make([]int, 1000000)[:3]`)

2. `defer` might slow you Down
: However consider the advantages.

3. strings are immutable
: Use [bytes.Buffer](https://golang.org/pkg/bytes/#Buffer) or [strings.Builder](https://golang.org/pkg/strings/#Builder)

4. Know when a goroutine is going to stop
: Avoid goroutine leaks. Use [context](https://golang.org/pkg/context/) for cancellation/timeouts.

5. `cgo` calls are expensive
: Group them together in one `cgo` call.

6. Channel can be slower than `sync.Mutex`
: However they are much easier to work with

7. Interface calls are more expensive the struct calls
: You can extract the value from the interface first. However it's less generic code.

8. Use `go run -gcflags=-m`
: You'll see what escapes to the heap.

### Reading

- [So you wanna go fast](https://www.slideshare.net/TylerTreat/so-you-wanna-go-fast-80300458)
- [Performance tuning workshop](https://github.com/davecheney/gophercon2018-performance-tuning-workshop/blob/master/6-tips-and-tricks/1-tips-and-tricks.md).
- [Quick look at some compiler optimization](http://www.golangbootcamp.com/book/tricks_and_tips#sec-compiler_optimizations)

## Non Go Specific

1. Know thy Hardware
: CPU affinity, CPU cache, memory, [latency numbers ...](https://people.eecs.berkeley.edu/~rcs/research/interactive_latency.html).
For example: [Cache-oblivious algorithms](https://en.wikipedia.org/wiki/Cache-oblivious_algorithm)

2. Algorithms & Data structures Rule
: They will usually give you much better performance than any other trick.

3. Include performance in your process
: Design & code reviews, run & compare benchmarks on CI ...
