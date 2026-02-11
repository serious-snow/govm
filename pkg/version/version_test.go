package version

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestVInfo_MarshalJSON(t *testing.T) {
	ss := []*Version{
		{
			Major: 1,
			Minor: 1,
			Patch: new(int),
			Beta:  true,
			VBeta: 1,
			RC:    false,
			VRC:   0,
		},
	}

	buf, err := json.Marshal(ss)
	fmt.Println(string(buf), err)
	s2 := make([]*Version, 0)
	fmt.Println(json.Unmarshal(buf, &s2))
	fmt.Println(s2)
}
