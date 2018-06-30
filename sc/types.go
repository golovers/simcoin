package sc

import (
	"encoding/hex"
)

type Hash []byte

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func StringToHash(v string) Hash {
	b, _ := hex.DecodeString(v)
	return b
}

type Address []byte

type PubKey []byte

type Stack struct {
	Values [][]byte
}

func NewStack() *Stack {
	return &Stack{Values: make([][]byte, 0)}
}

func (s *Stack) Push(v []byte) {
	s.Values = append(s.Values, v)
}

func (s *Stack) Pop() []byte {
	v := s.Values[len(s.Values)-1]
	s.Values = s.Values[:len(s.Values)-1]
	return v
}

func (s *Stack) Peak() []byte {
	return s.Values[len(s.Values)-1]
}
