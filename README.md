# 🔗 tcache [![Build Status](https://travis-ci.org/roobre/tcache.svg?branch=master)](https://travis-ci.org/roobre/tcache)  [![Maintainability](https://api.codeclimate.com/v1/badges/bc3394097831713f05fe/maintainability)](https://codeclimate.com/github/roobre/tcache/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/bc3394097831713f05fe/test_coverage)](https://codeclimate.com/github/roobre/tcache/test_coverage) 

tcache is a write-once, transactional cache specially suited for caching data which takes more time to be produced than to be re-requested.

Its transactional design ensures that subsequent accesses to the same key blocks until the key is filled by an ongoing transaction, waiting for data to be available and then serving it from cache.

## Use case

tcache defines an API intended to be used in environments where it is desired to cache data which takes a long time to produce, but can be requested multiple times in a short period of time.

The write-once transactional API of tcache makes easy to serve said requests easily and efficiently 

### Look & Feel

> Note: A more complete example in an HTTP server can be found [here](https://github.com/roobre/tcache/blob/master/example/answerServer.go) 

```go
func f() {
    var preciousData Data

    // Query from collection 
    err := cache.Access("mypreciousdata", 8*time.Hour, tcache.Handler{
        // Then function is executed if data is found in the cache
        // This function can be executed concurrently (as in sync.RWMutex.RLock())
        Then: func(cacheReader io.Reader) error {
            // Data found in cache
            return json.NewDecoder(cacheReader).Decode(preciousData)
        },
        // Else funciton is executed if data is not found in the cache,
        // and is used to populate it.
        // Accesses to the same key wait until the Else funciton finishes
        Else: func(cacheWriter io.Writer) error {
            // Think about the answer very carefully...
            preciousData = ComputeExpensiveStuff()
    
            // Store answer in cache
            return json.NewEncoder(cacheWriter).Encode(preciousData)
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println(preciousData)
}
```

## Behavior

tcache collections receive two functions, named `Then` and `Else`. `Then` is executed if an entry is found, and `Else` is executed if it is not.

These functions may also return errors, which modify the cache state and execution rules:

* If `Then` returns a non-nil error, the entry is flagged as *Invalid*, and the `Else` function will be executed after it
  - Other routines which were accessing the entry concurrently will not be aware of this change, which may or may not error and trigger their own execution of `Else`
  - If at least one of multiple concurrent `Then` executions fails, the entry is guaranteed to be invalidated and replaced by the entry produced by any of the `Else` function of the goroutines whose `Then` failed
* If `Else` returns a non-nil error, the entry is flagged as *Invalid* and the error same is returned by `Access`.

If an invalid entry is found when querying the cache, the behavior is the same as if said entry were not there.

`nil` is a valid handler for either `Then`, `Else`, or both. No action will be executed if the handler is nil, but existance, validity, and expiration checks will be performed, and `tcache.EntryMissingError` or `tcache.EntryInvalidatedError` will be returned accordingly.

## Stability

### Code

I originally developed this cache as a part of a bigger REST API project, for which being able to concurrently serve from cache expensive requests was a requirement. I later saw that it could be interesting to use this caching mechanism for other project, so tcache was born and extracted out.

While it is actively used by said project, it is for a hobby-like environment in which catastrophic failure or lack of performance are tolerated. I do not consider tcache to be stable or production-ready in any sense.

### API

Being early pre-alpha, I change the tcache API whenever I feel like it. I'm quite happy with its current shape, but there is absolutely no guarantee of this being its final form. 

## Implementing backends *(WIP)*

The `tcache` package manages the complex and specific concurrency tricks to provide transactional access to cache entries. The task of storing data is delegated to implementations of `tcache.Storage`, which tend to be simple.

> This section is a WIP and needs further expansion
