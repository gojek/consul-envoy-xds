package utils

import "github.com/gogo/protobuf/types"

func Uint32Value(value uint32) *types.UInt32Value {
	return &types.UInt32Value{Value: value}
}
