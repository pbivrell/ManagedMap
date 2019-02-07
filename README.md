# ManagedMap

## What is a ManagedMap
A managed map is a wrapper around a [go map](https://blog.golang.org/go-maps-in-action), that provideds thread safe operations as well as automatic key-value removal after a configureable timeout and/or number of reads.

## What can I put in a ManagedMap
The ManagedMap stores key-value pairs. The methods for Getting and Putting data into the map are defined with interface{} this implies that you can put any data type in as the key and value. This is partially true. The value can be __any__ go type. The key must be a type that can be compared with the == operator. If it is not the underlying go map will panic. For more reading see [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".

## Methods
Interactions with a managed map are done through the following methods.
* managedMap.Get(key interface{})
* managedMap.Put(key interface{}, value interface{})
* managedMap.Remove(key interface{})
* managedMap.Size()
* managedMap.Close()
* managedMap.PutCustom(key interface{}, value interface{}, conf Config)

## Example Useage
Get library with `go get github.com/pbivrell/ManagedMap`

#### Simple example
```go
import (
    "github.com/pbivrell/ManagedMap" 
    "fmt"
)

func main(){
    // Create new map
    m := ManagedMap.NewManagedMap()
    // Map must be closed before it can be garbage collected
    defer m.Close()
    // Insert new key-value pair    
    m.Put("First Key", 2)
    // Get value and existance of key in map
    value, has := m.Get("First Key")
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
    // Get value again. It will not exists because
    // the default number of accesses is 1
    value, has := m.Get("First Key")
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
}

```

#### Advanced Usage
``` go
import (
    "github.com/pbivrell/ManagedMap" 
    "fmt"
    "time"
)

func main(){
    // Create map with default timeout as 5 seconds and unlimited accesses
    m := ManagedMap.NewCustomManagedMap(ManagedMap.Conf(time.Second * 5, 0))
    defer m.Close()
    m.Put(12, nil)
    // Any number of access
    value, has := m.Get(12)
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
    value, has = m.Get(12)
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
    // Wait for time out
    time.Sleep(6 * time.Second)
    // This get will fail because of timeout
    value, has = m.Get(12)
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
    // Put item with single access overriding default
    m.PutCustom(true, 12.2, ManagedMap.Conf(time.Hour, 1))
    // Get it.
    value, has := m.Get(true)
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }
    // It should gone now
    value, has := m.Get(true)
    if has {
        fmt.Printf("Has item with value %v\n", value)
    }

}

```

## FAQ
__What does this error mean *'panic: runtime error: hash of unhashable type ...'*?__

As described [above](#What-can-I-put-in-a-ManagedMap) the ManagedMap allows you to __try__ to Put/Get any type of data. However the underlying data structure is a go map which only allows specific types into it namely only Boolean, Integer, Floating-point, Complex, String, Pointer, Channel, Interface, Struct, Array, and one other case. Inserting anything that is not one of these types will panic because of go's implementation of map. For more reading see [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
