// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package utils

import (
	"crypto/rand"
	"io"
	mrand "math/rand"

	"github.com/gofrs/uuid"
)

// GenerateDTMF generates a random string of 8 digits
func GenerateDTMF() (*string, error) {
	const size = 8
	table := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

	b := make([]byte, size)
	n, err := io.ReadAtLeast(rand.Reader, b, size)
	if n != size {
		return nil, err
	}

	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	result := string(b)
	return &result, nil
}

// RandomRange generates a random range in a particular range
// Reference: https://stackoverflow.com/a/36003006
func RandomRange(low, hi int) int {
	return low + mrand.Intn(hi-low)
}

// GenerateUUID generates a uuid string
func GenerateUUID() (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}
