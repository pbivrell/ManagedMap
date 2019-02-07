package ManagedMap

import (
    "testing"
    "time"
)

func TestPutGetRemoveSize(t *testing.T) {
    var tests = []struct{
        key interface{}
        value interface{}
        has bool
        remove bool
    }{
        {"A",1, true, false},
        {"B",1, false, false},
        {"B",1, true, false},
        {"C",1, true, false},
        {"D",1, true, true},
        {"E",1, true, false},
        {"F",1, true, true},
    }

    // Make a map with Unlimted access
    testMap := NewManagedMap()
    defer testMap.Close()
    size := 0
    for num, test := range tests {
        if test.has {
            testMap.Put( test.key, test.value)
            size++
        }
        value, has := testMap.Get(test.key)
        if has != test.has {
            t.Errorf("Test %d Failed: Inserted Key %v Value %v - Expected Exists: %v, Recieved Exists: %v\n",num +1, test.key, test.value, test.has, has)
        }
        if test.has && value != test.value{
            t.Errorf("Test %d Failed: Inserted Key %v Value %v - Expected Value: %v, Recieved Value: %v\n",num +1, test.key, test.value, test.value, value)
        }
        count := testMap.Size()
        if size != count {
            t.Errorf("Test %d Failed: Incorrect Size - Expected: %d, Recieved: %d\n",num+1, size, count)
        }
        if test.remove {
            testMap.Remove(test.key)
            size--
            _, has := testMap.Get(test.key)
            count := testMap.Size()
            if has || count != size {
                t.Errorf("Test %d Failed: Failed to remove item with key %v\n", num+1, test.key)
            }
        }
    }

}

func TestExpire(t *testing.T) {
    var tests = []struct {
        key      interface{}
        value    interface{}
        timeout  time.Duration
        wait     time.Duration
        has        bool
    }{
        {"apple", 1, 5 * time.Second, 0 * time.Second, true},
        {"apple", 1, 5 * time.Second, 10 * time.Second, false},
        {"apple", 1, 5 * time.Millisecond, 10 * time.Millisecond, false},
        {"apple", 1, 5 * time.Millisecond, 6 * time.Millisecond, false},
    }

    testMap := NewManagedMap()
    defer testMap.Close()
    for num, test := range tests {
        testMap.PutCustom(test.key, test.value, Config{test.timeout, 0})
        time.Sleep(test.wait)
        value, has:= testMap.Get(test.key)
        if has != test.has {
            t.Errorf("Test %d Failed: Inserted Key %v Value %v - Expected Exists: %v, Recieved Exists: %v\n",num +1, test.key, test.value, test.has, has)
        }else if test.has && value != test.value {
            t.Errorf("Test %d Failed: Inserted Key %v Value %v - Expected Value: %v, Recieved Value: %v\n",num+1, test.key, test.value, test.value, value)
        }
        testMap.Remove(test.key)
    }
}
