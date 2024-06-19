package codec

import (
  "errors"
  "fmt"
)

var (
  ErrEncode = errors.New("encoding error")
  ErrDecode = errors.New("decoding error")
)

func EncodeErr[T any](t T, msg string) error {
  return fmt.Errorf("%w: %T: %s", ErrEncode, t, msg)
}

func DecodeErr[T any](t T, msg string) error {
  return fmt.Errorf("%w: %T: %s", ErrDecode, t, msg)
}

type (
  Encoder interface {
    Encode() ([]byte, error)
  }

  Decoder interface {
    Decode(input []byte) (int, error)
  }

  Codec interface {
    Encoder
    Decoder
  }
)

