package base

// The most basic node, with no characteristics besides a Name
type Entry interface {
	Name() string
}

//////////////////////////////////////////////////////////////////////////////
// The rest of these interfaces define built-in entry types.
// Types are defined by the actions that they provide.
// In this way, golang interfaces are perfect analogues. What a concidence.
// Eventually types should have a representation in the tree,
//   that way new types can be introduced by the userspace.
// TODO: spec out what userspace types should be used for.

// A node that has named enumerable child Entries
type Folder interface {
	Entry
	Children() []string
	Fetch(name string) (entry Entry, ok bool)
	Put(name string, entry Entry) (ok bool)
}

// A special folder describing requirements that other nodes can be matched against
// Useful for defining structs for Functions to pass around (e.g. type system)
// Its own shape: type String, children []Shape, validate Function
type Shape interface {
	Folder
	Check(entry Entry) (ok bool)
}

// A node that can be invoked to map an Entry
type Function interface {
	Entry
	Invoke(input Entry) (output Entry)
}

// A node that has ordered enumerable children
type List interface {
	Entry
	// TODO: gotta be a better way to do this.
	Children() []Entry
}

// An immutable node which has a single read-only string value.
// Make a new node to represent a changed value.
type String interface {
	Entry
	Get() (value string)
}

// A node which has a known-size buffer of bytes
type File interface {
	Entry
	GetSize() uint
	Read(offset uint, numBytes uint) (data []byte)
	Write(offset uint, data []byte) (numBytes uint)
	Truncate() (ok bool)
}

// A node which acts as a FIFO pipe of foreign Entries
// (PS: should be usable for AWS SQS deliveries)
type Queue interface {
	Entry
	Push(value Entry) (ok bool)
	Close()
	Next() (value Entry, ok bool)    // waits for the next value
	TryNext() (value Entry, ok bool) // gets next value if one is waiting
}

// A node maintaining an append-only log of Entries.
// It's up to the implementation how long old entries are kept.
// (PS: should be compatible with AWS Kinesis)
type Log interface {
	Entry
	Append(value Entry) (ok bool)
	Close()
	Subscribe(opts Entry) (sub Queue) // writes values into dest
}

// TODO: this shouldn't be baked in, right?
// A node that can be subscribed to for updates
// Returns a folder representing the subscription
// Ideally supports backpressure and all that jazz
// nil functions won't be called
//type Observable interface {
//  Subscribe(onNext Function, onCompleted Function, onError Function) Folder
//}
