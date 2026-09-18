package shortid

import "strings"

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base = uint64(len(charset))

// converts a 64-bit integer into a Base62 string
func Encode(id uint64) string {
	if id == 0 {
		return string(charset[0])
	}

	var builder strings.Builder
	// maximum length of a base62 encoded uint64 is 11 characters
	builder.Grow(11)

	// keep dividing by 62 and mapping the remainder to a character
	var chars []byte
	for id > 0 {
		rem := id % base
		chars = append(chars, charset[rem])
		id = id / base
	}

	// Reverse the characters since we got them in least-significant order
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	builder.Write(chars)
	return builder.String()
}
