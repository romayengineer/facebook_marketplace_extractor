package main

import (
	"fmt"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

func DetectBestWithConfidence(data []byte) (string, error) {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err == nil && result != nil {
		if result.Confidence > 25 {
			return result.Charset, nil
		}
	}
	return "", fmt.Errorf("no encoding matches with condifence")
}

func GuessEncoding(data []byte) string {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err == nil && result != nil {
		LogDebug0("Encoding detected", "charset", result.Charset, "confidence", result.Confidence)
		return result.Charset
	}
	return "utf-8"
}

func DecodeWithEncoding(data []byte, charset string) ([]byte, error) {
	var enc encoding.Encoding
	switch strings.ToLower(charset) {
	case "windows-1252", "cp1252":
		enc = charmap.Windows1252
	case "iso-8859-1", "latin-1":
		enc = charmap.ISO8859_1
	case "utf-8", "utf8", "unknown":
		return data, nil
	default:
		return data, fmt.Errorf("unsupported encoding: %s", charset)
	}

	decoder := enc.NewDecoder()
	result, err := decoder.Bytes(data)
	if err != nil {
		return result, err
	}
	return result, nil
}
