package strutil

import "github.com/romayengineer/facebook_marketplace_extractor/core/logging"

func AreStringsEqual(s1, s2 string) bool {

	if s1 == s2 {
		logging.LogDebug0("AreStringsEqual", "Strings are equal")
		return true
	}

	minLen := min(len(s1), len(s2))

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			logging.LogDebug0("AreStringsEqual", "String difference found", "position", i, "byte1", s1[i], "byte2", s2[i])
			return false
		}
	}

	logging.LogDebug0("AreStringsEqual", "Strings differ in length", "len_s1", len(s1), "len_s2", len(s2))

	return false
}
