package model

type Node interface {
	Id() string
}

type NodeObject struct {
	id string
}

func (n *NodeObject) Id() string {
	return n.id
}

type NewNodeInput struct {
	Id string
}

func NewNode(input *NewNodeInput) *NodeObject {
	return &NodeObject{id: input.Id}
}
