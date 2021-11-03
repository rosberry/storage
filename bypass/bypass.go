package bypass

import (
	"errors"
)

type (
	Bypass struct {
		//
	}
)

var ErrMethodNotImplemented = errors.New("Method is not implemented")

var Instance = &Bypass{}

func New() *Bypass {
	return Instance
}

func (b *Bypass) Store(filePath, path string) (cLink string, err error) {
	return "", ErrMethodNotImplemented
}

func (b *Bypass) StoreByCLink(filePath, cLink string) (err error) {
	return ErrMethodNotImplemented
}

func (b *Bypass) GetURL(cLink string, options ...interface{}) (URL string) {
	return cLink
}

func (b *Bypass) GetCLink(path string) (cLink string) {
	return (path)
}

func (b *Bypass) Remove(cLink string) (err error) {
	return ErrMethodNotImplemented
}
