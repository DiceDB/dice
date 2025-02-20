package testutils

import "github.com/dicedb/dicedb-go/wire"

func ParseTestResponse(result *wire.Response) any {
	if result.Err != "" {
		return result.Err
	}
	switch v:= result.Value.(type){
	case *wire.Response_VStr:
		return v.VStr
	case *wire.Response_VInt:
		return v.VInt
	case *wire.Response_VBytes:
		return v.VBytes
	case *wire.Response_VFloat:
		return v.VFloat
	case *wire.Response_VNil:
		return v.VNil
	}
	return nil
}