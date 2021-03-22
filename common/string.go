package common

//DifferentElementStrings :
// find different elements between 2 strings
func DifferentElementStrings(src1, src2 []string) []string {
	res := []string{}
	m := map[string]bool{}
	lSrc := []string{}
	sSrc := []string{}
	if len(src1) >= len(src2) {
		lSrc = src1
		sSrc = src2
	} else {
		lSrc = src2
		sSrc = src1
	}
	for _, v := range sSrc {
		m[v] = true
	}
	for _, v := range lSrc {
		if !m[v] {
			res = append(res, v)
		}
	}
	return res
}

func InsertValueToSliceByIndex(list []string, value string, index int) []string {
	if len(list) == index { // nil or empty slice or after last element
		return append(list, value)
	}
	list = append(list[:index+1], list[index:]...) // index < len(a)
	list[index] = value
	return list
}

func DeepCopyString(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
