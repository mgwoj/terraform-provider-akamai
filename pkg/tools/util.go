// Package tools is where some legacy provider functions were dropped
package tools

import (
	"crypto/sha1"
	"encoding/hex"
	"log"

	"github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetSHAString returns a sha1 from the string
// TODO: utils should not exist, we should split this file into separate files, e.g. sha.go etc
func GetSHAString(rdata string) string {
	h := sha1.New()
	h.Write([]byte(rdata))

	sha1hashtest := hex.EncodeToString(h.Sum(nil))
	return sha1hashtest
}

// CreateNonce returns a random uuid string
// Deprecated: CreateNonce is deprecated, providers should use akactx.OperationID()
func CreateNonce() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Printf("[DEBUG] Generate Uuid failed %s", err)
		return ""
	}
	return uuid.String()
}

//Convert schema.Set to a slice of strings
func SetToStringSlice(s *schema.Set) []string {
	list := make([]string, s.Len())
	for i, v := range s.List() {
		list[i] = v.(string)
	}
	return list
}