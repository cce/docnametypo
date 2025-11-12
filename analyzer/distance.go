package analyzer

// passesDistanceGate ensures the distance match shares enough overlap.
func passesDistanceGate(doc, name string, dist int) bool {
	if dist <= 0 {
		return false
	}

	docLen := len(doc)
	nameLen := len(name)
	if docLen < minDocTokenLen+dist || nameLen < minDocTokenLen {
		return false
	}

	sharedPrefix := commonPrefixLength(doc, name)
	sharedSuffix := commonSuffixLength(doc, name)
	shared := sharedPrefix + sharedSuffix
	if shared > docLen {
		shared = docLen
	}

	required := docLen - dist
	if required < minDocTokenLen {
		required = minDocTokenLen
	}
	if shared >= required {
		return true
	}
	return docLen >= 2*minDocTokenLen && shared*2 >= docLen && docLen-shared <= dist
}

// commonPrefixLength returns the length of the shared prefix.
func commonPrefixLength(a, b string) int {
	limit := min(len(a), len(b))
	for i := 0; i < limit; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return limit
}

// commonSuffixLength returns the length of the shared suffix.
func commonSuffixLength(a, b string) int {
	ia := len(a) - 1
	ib := len(b) - 1
	count := 0
	for ia >= 0 && ib >= 0 {
		if a[ia] != b[ib] {
			break
		}
		count++
		ia--
		ib--
	}
	return count
}

// damerauLevenshtein computes the optimal string edit distance with transpositions.
// Simple O(len(a)*len(b)) DP; fine for short identifiers.
func damerauLevenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	na := len(ra)
	nb := len(rb)
	if na == 0 {
		return nb
	}
	if nb == 0 {
		return na
	}
	d := make([][]int, na+1)
	for i := 0; i <= na; i++ {
		d[i] = make([]int, nb+1)
		d[i][0] = i
	}
	for j := 0; j <= nb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= na; i++ {
		for j := 1; j <= nb; j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := d[i-1][j] + 1
			ins := d[i][j-1] + 1
			sub := d[i-1][j-1] + cost
			v := min(del, min(ins, sub))

			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				v = min(v, d[i-2][j-2]+1)
			}
			d[i][j] = v
		}
	}
	return d[na][nb]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
