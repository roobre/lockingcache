# 🔗 tcache

tcache is a write-once, transactional cache specially suited for caching data which takes more time to be produced than to be re-requested.

Its transactional design ensures that subsequent accesses to the same key blocks until the key is filled by an ongoing transaction, making data available and then be served from cache.

## Use case

tcache defines an API intended to be used in environments where it is desired to cache data which takes a long time to produce, but can be requested multiple times in a short period of time.

The write-once transactional API of tcache makes easy to serve said requests easily and efficiently 

### Look & Feel

> Note: A more complete example in an HTTP server can be found [here](https://github.com/roobre/tcache/blob/master/example/answerServer.go) 

```go
func f() {
    var preciousData Data

    // Query from collection 
    err := cache.From("myCollection").Access("mypreciousdata", 8*time.Hour, tcache.Handler{
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

## Stability

I have developed this project as an spin-off of a bigger HTTP server which needed a cache like this. I do not consider this code to be stable or production-ready in any sense.
