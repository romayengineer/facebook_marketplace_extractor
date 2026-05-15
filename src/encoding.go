package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

func GuessEncoding(data []byte) string {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	log.Printf("encoding %s confidence %02d", result.Charset, result.Confidence)
	if err == nil && result != nil {
		return result.Charset
	}
	return "utf-8"
}

func DecodeWithEncoding(data []byte, charset string) (string, error) {
	var enc encoding.Encoding
	switch strings.ToLower(charset) {
	case "windows-1252", "cp1252":
		enc = charmap.Windows1252
	case "iso-8859-1", "latin-1":
		enc = charmap.ISO8859_1
	case "utf-8", "utf8", "unknown":
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %s", charset)
	}

	decoder := enc.NewDecoder()
	result, err := decoder.Bytes(data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
