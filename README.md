# ManagedMap

## What is a ManagedMap
A managed map is a wrapper around a [go map](https://blog.golang.org/go-maps-in-action), that provideds thread safe operations as well as automatic key-value removal after a configureable timeout and/or number of reads.

## What can I put in a ManagedMap
The ManagedMap stores key-value pairs. The methods for Getting and Putting data into the map are defined with interface{} this implies that you can put any data type in as the key and value. This is partially true. The value can be __any__ go type. The key must be a type that can be compared with the == operator. If it is not the underlying go map will panic. For more reading see [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".

## Methods
Interactions with a managed map are done through the following methods.
* managedMap.Get(key interface{})
* managedMap.Put(key interface{}, value interface{})
* managedMap.Remove(key ineterface{})
* managedMap.Size()
* managedMap.Close()
* managedMap.PutCustom()

## Why would I use a ManagedMap


## Useage

## FAQ
What does this error mean 'panic: runtime error: hash of unhashable type ...'?
As described [above](#What-can-I-put-in-a-ManagedMap) the ManagedMap allows you to __try__ to Put/Get any type of data. However the underlying data structure is a go map which only allows specific types into it namely only Boolean, Integer, Floating-point, Complex, String, Pointer, Channel, Interface, Struct, Array, and one other case. Inserting anything that is not one of these types will panic because of go's implementation of map. For more reading see [Go maps in action](https://blog.golang.org/go-maps-in-action) the section about "Key types".
