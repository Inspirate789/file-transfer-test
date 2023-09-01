package generator

import (
	"io"
	"math/rand"
	"os"
	"time"
)

func GenerateFile(filename string, size int64) (err error) {
	reader := rand.New(rand.NewSource(time.Now().UnixNano()))
	var file *os.File
	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(file)

	_, err = io.CopyN(file, reader, size)
	if err != nil {
		return err
	}

	return nil
}
