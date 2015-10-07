// Built-in map doesn't support concurrent.
// This is concurrent map using channel, without mutex.
// https://gist.github.com/jaehue/5d1aaf76d082f98e8dc0 에서 가져옴 (key string 타입 -> int64 타입)
// TODO Generic 하게 사용할 방법을 찾아야 함..code generate 를 하거나 밑에 방법처럼?
// http://blog.burntsushi.net/type-parametric-functions-golang/ (느린거 같음..)
package gsutil


// A thread safe map(type: `map[int64]interface{}`).
// This using channel, not mutex.
type SharedMap interface {
	// Sets the given value under the specified key
	Set(k int64, v interface{})

	// Retrieve an item from map under given key.
	Get(k int64) (interface{}, bool)

	// Remove an item from the map.
	Remove(k int64)

	// Return the number of item within the map.
	Count() int

	Map() map[int64]interface{}

	// Return all the keys or a subset of the keys of an Map
	GetKeys() []int64
}

type sharedMap struct {
	m map[int64]interface{}
	c chan command
}

type command struct {
	action int
	key    int64
	value  interface{}
	result chan<- interface{}
}

const (
	set = iota
	get
	remove
	count
	keys
)

func (sm sharedMap) Map() map[int64]interface{} {
	return sm.m
}

// Sets the given value under the specified key
func (sm sharedMap) Set(k int64, v interface{}) {
	sm.c <- command{action: set, key: k, value: v}
}

// Retrieve an item from map under given key.
func (sm sharedMap) Get(k int64) (interface{}, bool) {
	callback := make(chan interface{})
	sm.c <- command{action: get, key: k, result: callback}
	result := (<-callback).([2]interface{})
	return result[0], result[1].(bool)
}

// Remove an item from the map.
func (sm sharedMap) Remove(k int64) {
	sm.c <- command{action: remove, key: k}
}

// Return the number of item within the map.
func (sm sharedMap) Count() int {
	callback := make(chan interface{})
	sm.c <- command{action: count, result: callback}
	return (<-callback).(int)
}

// Return all the keys or a subset of the keys of an Map (추가함)
func (sm sharedMap) GetKeys() []int64 {
	callback := make(chan interface{})
	sm.c <- command{action: keys, result: callback}
	return (<-callback).([]int64)
}

func (sm sharedMap) run() {
	for cmd := range sm.c {
		switch cmd.action {
		case set:
			sm.m[cmd.key] = cmd.value
		case get:
			v, ok := sm.m[cmd.key]
			cmd.result <- [2]interface{}{v, ok}
		case remove:
			delete(sm.m, cmd.key)
		case count:
			cmd.result <- len(sm.m)
		case keys:
			keys := make([]int64, 0, 1024)
			for key, _ := range sm.m {
				keys = append(keys, key)
			}
			cmd.result <- keys
		}
	}
}

// Create a new shared map.
func NewSMap() SharedMap {
	sm := sharedMap{
		m: make(map[int64]interface{}),
		c: make(chan command),
	}
	go sm.run()
	return sm
}


