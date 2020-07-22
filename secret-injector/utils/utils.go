package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"unicode"
)

// Help us identify the string has no unicode
// A string that is not base64 encoded will return false
func checkUnicode(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// Encode the value of the secret provided by vault
func EncodeValue(m map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})
	for k, v := range m {
		// Base64 encoded should return `true` in this function otherwise encode the string to base64
		ok, err := regexp.MatchString("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", v.(string))
		if ok {
			decode, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				if _, ok := err.(base64.CorruptInputError); ok {
					panic("\nbase64 input is corrupt, check your secret")
				}
			}
			ok := checkUnicode(string(decode))
			if !ok {
				value := base64.StdEncoding.EncodeToString([]byte(v.(string)))
				payload[k] = value
			}
			if ok {
				payload[k] = v
			}
		}
		if !ok {
			value := base64.StdEncoding.EncodeToString([]byte(v.(string)))
			payload[k] = value
		}
		if err != nil {
			fmt.Println("encountered an error on regexp")
		}
	}
	return payload
}

