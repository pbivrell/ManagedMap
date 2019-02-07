package ManagedMap
// ManagedMap is a wrapper around a go map that provides synchronization 
// via a sync.RWMutex. It also provides configurable automatic key-value pair 
// removal based on a timeout and/or a number of accesses. The underlying
// datastruct is intentionally unexported all interactions should be done
// via the Methods provided in this package

import (
    "time"
    "sync"
    "sync/atomic"
    "math"
)

const (
    DefaultTimeout = 24 * time.Hour
    DefaultAccessCount = 1
)

// Config is the struct that is used to Config the timeout and accessCount
// of both the default values for all map items as well as individual map
// items. The value '0' for either Timeout or AccessCount is interpreted
// as infinite and the maxium value of the respective types will be used
// in place of infinity. 
type Config struct {
    Timeout     time.Duration
    AccessCount uint64
}

// item is a private struct that manages the internal value of the map.
// This manages the timer, the accesses, the data, and a closed channel.
// item is unexported but allows the user to use any data they desire to be
// stored in the map.
type item struct {
    timer *time.Timer
    accessRemaining uint64
    data interface{}
    removed chan bool
}

// managedMap is a private struct that manages the internals of the managedMap
// the data access paterns are very particular for this data structure. For that
// reason it and its memebers are unexported. Users of this structure are required
// to make use of provided methods.
type managedMap struct {
    default_timeout time.Duration
    default_access  uint64
    m map[interface{}] *item
    lock               *sync.RWMutex
}

// NewManagedMap returns a pointer to a managedMap with the default timeout and accessCount
// as defined by the DefaultTimeout and DefaultAccessCount constants. 
func NewManagedMap() *managedMap {
    return NewCustomManagedMap(Config{Timeout: DefaultTimeout, AccessCount: DefaultAccessCount})
}

// NewCustomManagedMap returns a pointer to a managedMap with the timeout and accessCount
// defined by the passed Config struct.
func NewCustomManagedMap(conf Config) *managedMap {
    m := make(map[interface{}] *item)
    lock := &sync.RWMutex{}
    return &managedMap{
        default_timeout: conf.Timeout,
        default_access: conf.AccessCount,
        m: m,
        lock: lock,
    }
}

// Get is a method of a managedMap that returns the value associated with
// the passed key and a boolean representing whether or not it exists. 
// A key does not exist if it was never inserted or has been removed by timeout
// or exceeding accesses limit. Get will always panic when called after the Close
// method has been called. The key must be a type that can be compared with the == operator. 
// If it is not the underlying go map will panic. For more reading see 
// [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
func (t *managedMap) Get(key interface{}) (interface{}, bool) {
    t.lock.RLock()
    defer t.lock.RUnlock()
    // Panic if managedMap is closed
    t.closed()
    // Check if the item exists. Return if it doesn't
    item, has := t.m[key]
    if !has {
        return nil, false
    }
    // Atomically read the number of access remaining
    accesses := atomic.LoadUint64(&item.accessRemaining)
    // If accesses remaining is 0 that means this has already
    // been read more then its alotted amount of times. Its possible
    // that the element is not quite deleted yet here so we prented that
    // it has already been delete.
    if accesses < 1 {
        return nil, false
    }
    // Hack to add negative 1 to a unit64. This is safe because at this point
    // accesses is a positive value greater then 1.
    negative1 := int64(-1)
    atomic.StoreUint64(&item.accessRemaining, accesses + uint64(negative1))
    // If this is the last access we can use the removed channel to delete the
    // key. This is done in a goroutine so that the Get call does not block 
    // to aquire the write lock.
    if accesses == 1 {
        go func(t *managedMap, removed chan bool) {
            t.lock.Lock()
            defer t.lock.Unlock()
            removed <- true
            <-removed
        }(t, item.removed)
    }
    return item.data, true
}

// Put is a method of a managedMap that allows the user to insert a key-value pair.
// Calling Put with a key that already exists will update the value but
// will not alter the timer or the access count. The Put will always panic when called
// after the Close method has been called. The key must be a type that can be compared 
// with the == operator. If it is not the underlying go map will panic. For more reading see 
// [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
func (t *managedMap) Put(key, value interface{}) {
    t.PutCustom(key,value, Config{ t.default_timeout, t.default_access })
}

// Remove is a method of a managedMap that allows the user to remove a key and its
// associated data from the map specifically the timer and access counts will cleared.
// Remove method will panic when called after the Close method has been called. The key 
// must be a type that can be compared with the == operator. If it is not the underlying 
// go map will panic. For more reading see 
// [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
func (t *managedMap) Remove(key interface{}) {
    t.lock.Lock()
    defer t.lock.Unlock()
    // Panic if managedMap is closed
    t.closed()
    value, has := t.m[key]
    if has {
        value.removed <- true
        <-value.removed
    }
}

// Size is a method of a managedMap that will return the number of items
// stored in the map. Size will panic when called after the Close method 
// has been called.
func (t *managedMap) Size() int {
    t.lock.RLock()
    defer t.lock.RUnlock()
    // Panic if managedMap is closed
    t.closed()
    return len(t.m)
}


// Close is a method of a managedMap that cleans a ManagedMap. Any underlying data is set to
// nil and all Goroutines are stopped. 
func (t *managedMap) Close() {
    t.lock.Lock()
    defer t.lock.Unlock()
    // Panic if managedMap is closed
    t.closed()
    for _, v := range t.m {
        v.removed <- true
        <-v.removed
    }
    t.m = nil
}

// PutCustom is a method of a managedMap that allows the user to insert a key-value
// pair with custom values for timeout and access count in the form of a Config struct.
// Calling PutCustom with a key that already exists will update the value but
// will not alter the timer or the access count. PutCustom will always panic when called
// after the Close method has been called. The key must be a type that can be compared with the == operator. 
// If it is not the underlying go map will panic. For more reading see 
// [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
func (t *managedMap) PutCustom(key, value interface{}, config Config) {
    // Update only value if it already exists in the map
    t.lock.RLock()
    // Panic if managedMap is closed
    t.closed()
    // Update value if it already exists
    if v, has := t.m[key]; has {
        v.data = value
        t.lock.RUnlock()
        return
    }
    t.lock.RUnlock()
    // '0' as a config value implies infinite. We use math make value to suplement infinity.
    if config.Timeout == 0 {
        config.Timeout = math.MaxInt64
    }
    if config.AccessCount == 0 {
        config.AccessCount = math.MaxUint64
    }
    // Create a new map item
    timer := time.NewTimer(config.Timeout)
    removed := make(chan bool)
    item := &item{
        timer: timer,
        accessRemaining: config.AccessCount,
        data: value,
        removed: removed,
    }
    // Grab lock as writer update the map
    t.lock.Lock()
    defer t.lock.Unlock()
    t.m[key] = item
    // Spawn goroutine which will manage the newly created map item. This routine will
    // block until the timer expires or the items is removed. 
    go func(timer *time.Timer, t *managedMap, key interface{}, removed chan bool) {
        select {
            // Waits on the removed channel. If the removed channel recieves a value
            // we can safely delete the the key from the map because the senders aquire
            // a write lock on the data before sending data. The sender holds this lock
            // until we sender back on the same channel
        case <-removed:
            delete(t.m, key)
            removed <- true
            // Waits on the timer channel. If the timer has expired we need to aquire
            // the write lock before we can delete the data.
        case <-timer.C:
            t.lock.Lock()
            defer t.lock.Unlock()
            delete(t.m, key)
        }
    }(timer, t, key, removed)
}

// closed is a private method of a managedMap that panics if the Close method 
// has been called. This is used internally to ensure that no methods are 
// called after the data structure is closed. 
func (t *managedMap) closed(){
    if t.m == nil {
        panic("Could not perform Close on a closed managedMap")
    }
}

