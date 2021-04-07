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

// InsertValueToSliceByIndex insert a value into list, shift current right current value of that index
func InsertValueToSliceByIndex(list []string, value string, index int) []string {
	// TODO: @tin what if index > len(list) or index < 0 (maybe bug in calling function)
	if len(list) < index || index < 0 {
		msg := fmt.Sprintf("try to insert at index %+v but list length is %+v", index, len(list))
		panic(msg)
	}
	if len(list) == index { // nil or empty slice or after last element
		return append(list, value)
	}
	// TODO: @tin should make a new slice instead of using append
	//list = append(list[:index+1], list[index:]...) // index < len(a)
	//list[index] = value

	newList := make([]string, 0, 0)
	newList = append(newList, list[:index]...)
	newList = append(newList, value)
	newList = append(newList, list[index:]...)

	return newList
}
