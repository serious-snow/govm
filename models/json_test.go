package models

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestVInfo_MarshalJSON(t *testing.T) {
	var ss = []*Version{
		&Version{
			V1:    1,
			V2:    1,
			V3:    2,
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
