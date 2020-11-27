package tcache

import (
	"bytes"
	"io"
)

type NilStorage struct{}

func (NilStorage) Get(key string) Accessor {
	return NilAcessor{}
}

func (NilStorage) Delete(key string) {

}

type NilAcessor struct{}

func (NilAcessor) Reader() (io.Reader, error) {
	return bytes.NewReader(nil), nil
}

func (NilAcessor) Writer() (io.Writer, error) {
	return &bytes.Buffer{}, nil
}
