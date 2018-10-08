package server

import "context"

// MeshxService between meshx client agent and meshx server agent.
type MeshxService struct {
	agent *Agent
}

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

// Invoke is a registered service.
func (s *MeshxService) Invoke(ctx context.Context, args *Args, reply *Reply) error {
	resp, err := s.agent.adapter.Send(args.Frame)
	if err != nil {
		return err
	}
	*reply = Reply{Data: resp}
	return nil
}
