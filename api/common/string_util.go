package common

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	storedCode TextCode
	seed       = rand.NewSource(time.Now().UnixNano())
	randomizer = rand.New(seed)
)

type TextCode struct {
	Kind   string `json:"kind,omitempty"`
	UserID string `json:"userId"`
	Code   string `json:"code"`
}

func (c TextCode) String() string {
	return ObjToString(c)
}

func JoinErrors(errs []error) error {
	var sb bytes.Buffer
	for _, err := range errs {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return errors.New(sb.String())
}

func ObjToString(o interface{}) string {
	b, _ := json.Marshal(o)
	return string(b)
}

func GenerateTextCode() string {
	letters := strings.SplitN("0123456789", "", -1)
	sb := bytes.Buffer{}
	max := 6
	for i := 0; i < max; i++ {
		sb.WriteString(letters[randomizer.Intn(len(letters))])
	}
	return sb.String()
}

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return string(b), nil
}
