package Search

import (
	"strings"
)

func queryV1toV2(in string) (out string) {
	quotedQuery := `"` + in + `"`
	quotedQuery = strings.Replace(quotedQuery, ` AND `, `" AND "`, -1)
	quotedQuery = strings.Replace(quotedQuery, ` OR `, `" OR "`, -1)
	quotedQuery = strings.Replace(quotedQuery, ` NOT `, `" NOT "`, -1)
	qBits := strings.Split(quotedQuery, " NOT ")
	for i := range qBits {
		qqBits := strings.Split(qBits[i], " OR ")
		if len(qqBits) > 1 {
			for ii := range qqBits {
				qqBits[ii] = "(" + qqBits[ii] + ")"
			}
		}
		qBits[i] = "(" + strings.Join(qqBits, " OR ") + ")"
	}
	quotedQuery = strings.Join(qBits, " NOT ")
	return quotedQuery
}
