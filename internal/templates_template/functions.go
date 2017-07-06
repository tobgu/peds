package templates_template

import (
	"fmt"
	"hash/fnv"
)

func xhash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func xgenericHashFunc(x interface{}) uint64 {
	return hash(fmt.Sprintf("%v", x))
}
