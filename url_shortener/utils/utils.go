// util functions like generating random string, validating url etc
package utils

import (
	"encoding/hex"
	"crypto/rand"
	"strings"
	"net/url"
)


func GenerateShortCode(length int) (string, error){
	// create a byte map
	bytes := make([]byte, length/2)
	if _,err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func IsValidUrl(urlStr string) bool {
	
	// Still a string, need to parse this as a url
	u, err := url.Parse(urlStr)

	return err==nil && u.Host!="" && u.Scheme!=""
}

func NormalizeUrl(urlStr string) string {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://"){
		return "https://"+urlStr
	}
	return urlStr
}