package Search

import (
	"github.com/300brand/searchquery"
	"testing"
)

var queryTests = []struct {
	In  string
	Out string
	SQ  string
}{
	{
		`CDW OR CDWG NOT collision damage waiver`,
		`(("CDW") OR ("CDWG")) NOT ("collision damage waiver")`,
		`+((+:"CDW") (+:"CDWG")) -(+:"collision damage waiver")`,
	},
}

func TestV1toV2(t *testing.T) {
	for i, test := range queryTests {
		out := queryV1toV2(test.In)
		if test.Out != out {
			t.Errorf("[%d] V1toV2 Expected: %s", i, test.Out)
			t.Errorf("[%d] V1toV2 Got:      %s", i, out)
			continue
		}
		q, err := searchquery.ParseGreedy(out)
		if err != nil {
			t.Errorf("[%d] Error parsing: %s", i, err)
			continue
		}
		if q.String() != test.SQ {
			t.Errorf("[%d] searchquery Expected: %s", i, test.SQ)
			t.Errorf("[%d] searchquery Got:      %s", i, q)
		}
	}
}
