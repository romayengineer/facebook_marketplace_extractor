package main

import (
	"log"
)

func AreStringsEqual(s1, s2 string) bool {

	if s1 == s2 {
		log.Printf("Strings are equal\n")
		return true
	}

	minLen := min(len(s1), len(s2))

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			log.Printf("First difference at position %d: byte %d (0x%02x) vs byte %d (0x%02x)\n", i, s1[i], s1[i], s2[i], s2[i])
			log.Printf("Context s1: %s\n", s1[max(0, i-10):min(len(s1), i+30)])
			log.Printf("Context s2: %s\n", s2[max(0, i-10):min(len(s2), i+30)])
			return false
		}
	}

	log.Printf("Strings differ in length: s1 = %d, s2 = %d\n", len(s1), len(s2))

	return false
}
