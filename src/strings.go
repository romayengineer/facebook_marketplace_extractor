package main

func AreStringsEqual(s1, s2 string) bool {

	if s1 == s2 {
		LogDebug0("Strings are equal")
		return true
	}

	minLen := min(len(s1), len(s2))

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			LogDebug0("String difference found", "position", i, "byte1", s1[i], "byte2", s2[i])
			return false
		}
	}

	LogDebug0("Strings differ in length", "len_s1", len(s1), "len_s2", len(s2))

	return false
}
