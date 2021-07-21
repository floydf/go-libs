package lib

import (
	"log"
	"testing"
)

type SampleObject struct {
	Aaa, Bbb, Ccc string
	Ddd, Eee, Fff int
}

func (so SampleObject) String() string {
	return Jsonify(so)
}

func TestJsonify(t *testing.T) {
	log.Printf("testing")

	so := SampleObject{
		Aaa: "AAA",
		Bbb: "BBB",
		Ccc: "CCC",
		Ddd: 1,
		Eee: 2,
		Fff: 3,
	}

	log.Println(Jsonify(so))
	log.Println(so)
}
