package client

// Args is the service argument.
type Args struct {
	//Seq     uint64
	Service string
	Method  string
	Frame   []byte
}

// Reply is the service reply.
type Reply struct {
	//Seq  uint64
	Data []byte
}
