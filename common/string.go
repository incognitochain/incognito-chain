package common

func DeepCopyString(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
