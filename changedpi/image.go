package changedpi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func encode(src []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return dst
}

// decode decodes the provided BASE64 encoded data.
func decode(data []byte) ([]byte, error) {
	src := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(src, data)
	if err != nil {
		return nil, errors.New("base64.StdEncoding.Decode failed: " + err.Error())
	}
	return src[:n], nil
}

// SaveImage saves a BASE64 encoded image string to an output file.
func SaveImage(output, base64Image string) error {
	// Split the base64 header and data
	sp := strings.Split(base64Image, "base64,")
	if len(sp) != 2 {
		return errors.New("base64 format error")
	}
	base64Str := sp[1]

	// Decode the Base64 string
	data, err := decode([]byte(base64Str))
	if err != nil {
		return err
	}

	// Create and write to the output file
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the decoded image data to the file
	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

// EncodeFileString encodes file content of `path` using BASE64 algorithms.
func EncodeFileString(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		err = errors.New(fmt.Sprintf(`os.ReadFile failed for filename "%s"`, filename))
		return "", err
	}
	return string(encode(content)), nil
}

func GetBase64Image(filename string) (string, error) {
	suffix := strings.ToLower(path.Ext(filename))
	input, err := EncodeFileString(filename)
	if err != nil {
		return "", err
	}

	if suffix == ".png" {
		return "data:image/png;base64," + input, nil
	}

	if suffix == ".jpg" || suffix == ".jpeg" {
		return "data:image/jpeg;base64," + input, nil
	}

	return "", nil
}

func ChangeDpiByPath(path string, dpi int) (string, error) {
	baseStr, err := GetBase64Image(path)
	if err != nil {
		return "", err
	}

	output, err := ChangeDpi(baseStr, dpi)
	if err != nil {
		return "", nil
	}
	return output, nil
}
