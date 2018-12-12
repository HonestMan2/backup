package blkreward

import (
	"encoding/json"
	"fmt"

	"testing"
)

type ASD struct {
	A int
	B int
}

func TestNew(t *testing.T) {

	ans := ASD{
		A: 1,
		B: 2,
	}

	//p := &ans

	data, err := json.Marshal(&ans)
	fmt.Println(string(data), err)

	pp := new(ASD)
	err = json.Unmarshal(data, pp)
	fmt.Println(pp, err)
	fmt.Println(pp.A)

}
