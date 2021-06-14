package committeestate

import "fmt"

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

// insertValueToSliceByIndex insert a value into list, shift current right current value of that index
func insertValueToSliceByIndex(list []string, value string, index int) []string {
	if index > len(list) || index < 0 {
		msg := fmt.Sprintf("try to insert at index %+v but list length is %+v", index, len(list))
		panic(msg)
	}

	// nil or empty slice or after last element
	if len(list) == index {
		return append(list, value)
	}

	newList := make([]string, 0, 0)
	newList = append(newList, list[:index]...)
	newList = append(newList, value)
	newList = append(newList, list[index:]...)

	return newList
}
