package sdk

import (
	"fmt"

	"github.com/bolaxy/common"
	"github.com/bolaxy/common/hexutil"
	"github.com/bolaxy/crypto"
)

type Type string
type Kind uint8
type Callback func(*Event)
type Cancel func()

const (
	fireOnce Kind = iota
	fireAlways
)

type Result struct {
	Success         bool
	ContractAddress *common.Address
	IsLog           bool
	Data            []byte
	Topics          []common.Hash
}

type Event struct {
	eventType Type
	value     interface{}
}

func GenLogType(contractAddr common.Address, eventSig common.Hash) Type {
	return Type(hexutil.Encode(crypto.Keccak256(contractAddr.Bytes(), eventSig.Bytes())))
}

func GenHashType(txhash common.Hash) Type {
	return Type(txhash.String())
}

func NewEvent(evtType Type, value interface{}) *Event {
	return &Event{
		eventType: evtType,
		value:     value,
	}
}

func (e *Event) GetValue() interface{} {
	return e.value
}

func (e *Event) GetType() Type {
	return e.eventType
}

func NewEventEmitter(bufSize int) *Emitter {
	ee := &Emitter{
		bufSize:      bufSize,
		events:       make(map[uint64]map[Type]Kind),
		subscriber:   make(map[Type]map[uint64]chan<- *Event),
		observer:     make(chan *observer),
		notify:       make(chan *Event), // 如果设置了buf，在cancel阶段会出现竞争问题，然后回调函数中会多次调用cancel
		cancellation: make(chan uint64),
	}

	go ee.run()
	return ee
}

type Emitter struct {
	counter      uint64
	bufSize      int
	events       map[uint64]map[Type]Kind
	subscriber   map[Type]map[uint64]chan<- *Event
	observer     chan *observer
	notify       chan *Event
	cancellation chan uint64
}

func (ee *Emitter) run() {
	for {
		select {
		case observer := <-ee.observer:
			ee.counter += 1
			for _, et := range observer.eventTypes {
				fmt.Printf("eventType: %s, counter: %d\n", et, ee.counter)
				item, ok := ee.events[ee.counter]
				if !ok {
					item = make(map[Type]Kind)
					ee.events[ee.counter] = item
				}

				item[et] = observer.kind

				cb, ok := ee.subscriber[et]
				if !ok {
					cb = make(map[uint64]chan<- *Event)
					ee.subscriber[et] = cb
				}

				input := make(chan *Event)
				newObservable(ee.counter, input, ee.bufSize, observer.fn)
				cb[ee.counter] = input
			}
			observer.signal <- ee.counter
		case e := <-ee.notify:
			if item, ok := ee.subscriber[e.GetType()]; ok {
				for id, ch := range item {
					fmt.Printf("%s -> %d\n", e.GetType(), id)
					kind := ee.events[id][e.GetType()]
					ch <- e
					if kind == fireOnce {
						fmt.Printf("fireonce ...\n")
						delete(ee.events, id)
						delete(ee.subscriber[e.GetType()], id)
						close(ch)
					}
				}
			}
		case id := <-ee.cancellation:
			if items, ok := ee.events[id]; ok {
				delete(ee.events, id)
				for et := range items {
					close(ee.subscriber[et][id])
					delete(ee.subscriber[et], id)
				}
			}
		}
	}
}

func (ee *Emitter) On(fn Callback, eventType ...Type) (cancel Cancel) {
	return ee.subscribe(fireAlways, fn, eventType...)
}

func (ee *Emitter) Once(fn Callback, eventType ...Type) (cancel Cancel) {
	return ee.subscribe(fireOnce, fn, eventType...)
}

func (ee *Emitter) subscribe(kind Kind, fn Callback, eventType ...Type) (cancel Cancel) {
	if len(eventType) == 0 {
		return func() {}
	}
	ch := make(chan uint64, 1)
	ee.observer <- &observer{
		kind:       kind,
		fn:         fn,
		eventTypes: eventType,
		signal:     ch,
	}

	id := <-ch
	return func() {
		ee.cancellation <- id
	}
}

func (ee *Emitter) Emit(event *Event) {
	ee.notify <- event
}

type observer struct {
	kind       Kind
	fn         Callback
	eventTypes []Type
	signal     chan<- uint64
}

type observable struct {
	input  <-chan *Event
	output chan *Event
	fn     Callback
	index  uint64
}

// output must buffered channel
func newObservable(index uint64, input <-chan *Event, bufSize int, fn Callback) *observable {
	if bufSize <= 0 {
		bufSize = 128
	}

	output := make(chan *Event, bufSize)
	r := &observable{
		input:  input,
		output: output,
		fn:     fn,
		index:  index,
	}

	go r.consume()
	go r.run()
	return r
}

func (r *observable) run() {
	for v := range r.input {
		select {
		case r.output <- v:
		default:
			x := <-r.output
			fmt.Printf("(%d)Warning: drop event: %v\n", r.index, x)
			r.output <- v
		}
	}

	fmt.Printf("(%d) quit.\n", r.index)
	close(r.output)
}

func (r *observable) consume() {
	for e := range r.output {
		fmt.Printf("(%d)consume 1----> %v\n", r.index, e)
		r.fn(e)
		fmt.Printf("(%d)consume 2----> %v\n", r.index, e)
	}
}
