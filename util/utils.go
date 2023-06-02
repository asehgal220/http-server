package util

func FileIndex(allFiles []string, file string) int8 {
	for i := 0; i < len(allFiles); i++ {
		if allFiles[i] == file {
			return int8(i)
		}
	}
	return -1
}
