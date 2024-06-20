package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	"github.com/DvdSpijker/GoBroker/codec"
)

var supportedSizes = []int{1, 2, 4}

const MaxStringLength = 65535

const (
  QoS0 QoS = 0b00
  QoS1 QoS = 0b01
  QoS2 QoS = 0b10
)

type (
  QoS byte

	UnsignedInt struct {
		Value uint32
		Size  int
	}

	UtfString struct {
		Str string
	}

	UtfStringPair struct {
		Name  UtfString
		Value UtfString
	}

	BinaryData struct {
		Data []byte
	}

	VariableByteInteger struct {
		Value int32
	}
)


func SetBit(b byte, pos uint8) byte {
	b |= (1 << pos)
	return b
}

func ClearBit(b byte, pos uint8) byte {
	mask := byte(^(1 << pos))
	b &= mask
	return b
}

func GetBit(b byte, pos uint8) bool {
	return b&(1<<pos) > 0
}

// Encodes an unsigned integer as:
// | MSB | .. | LSB |
func (integer *UnsignedInt) Encode() ([]byte, error) {
	if !slices.Contains(supportedSizes, integer.Size) {
		return []byte{}, codec.EncodeErr(
			integer,
			fmt.Sprintf("unsupported size: %d", integer.Size))
	}
	encoded := make([]byte, 0, integer.Size)

	err := binary.Write(bytes.NewBuffer(encoded), binary.BigEndian, &integer.Value)
	if err != nil {
		return []byte{}, errors.Join(codec.EncodeErr(integer, "buffer write"), err)
	}

	return encoded, nil
}

// Size must be set before decoding.
func (integer *UnsignedInt) Decode(input []byte) (int, error) {
	if !slices.Contains(supportedSizes, integer.Size) {
		return 0, codec.DecodeErr(
			integer,
			fmt.Sprintf("unsupported size: %d", integer.Size))
	}

  input = input[:integer.Size]
  fmt.Println("unsigned int input", input)
  switch integer.Size {
  case 1:
    integer.Value = uint32(input[0])
  case 2:
    integer.Value = uint32(binary.BigEndian.Uint16(input))
  case 4:
    integer.Value = uint32(binary.BigEndian.Uint32(input))
}

	return integer.Size, nil
}

// Encodes a UTF string as:
// | len MSB | len LSB | UTF-8 data |
func (utfString *UtfString) Encode() ([]byte, error) {
	if len(utfString.Str) > MaxStringLength {
		return []byte{}, codec.EncodeErr(
			utfString,
			fmt.Sprintf("unsupported string length: %d", len(utfString.Str)))
	}
	encoded := make([]byte, 0, len(utfString.Str)+2)

	length := UnsignedInt{Value: uint32(len(utfString.Str))}
	encLength, err := length.Encode()
	if err != nil {
		return []byte{},
			errors.Join(codec.EncodeErr(utfString, "length encoding error"), err)
	}

	copy(encoded, encLength)
	copy(encoded[length.Value:], []byte(utfString.Str))

	return encoded, nil
}

func (utfString *UtfString) Decode(input []byte) (int, error) {
	if len(input) < 2 {
		return 0, codec.DecodeErr(utfString, "input must be at least 2 bytes")
	}
	// length := int(input[0])
	// length = length << 8
	// length += int(input[1])
	length := binary.BigEndian.Uint16(input[:2])
	fmt.Println("string length:", length)

	input = input[2:]
	utfString.Str = ""
	for i := 0; i < int(length); i++ {
		utfString.Str += string(input[i])
	}

	return int(length)+ 2, nil
}

func (utfString *UtfString) String() string {
	return utfString.Str
}

func (utfStringPair *UtfStringPair) Encode() ([]byte, error) {
	encoded := []byte{}
	bytes, err := utfStringPair.Name.Encode()
	if err != nil {
		// TODO: proper error
		return []byte{}, err
	}

	encoded = append(encoded, bytes...)

	bytes, err = utfStringPair.Value.Encode()
	if err != nil {
		// TODO: proper error
		return []byte{}, err
	}

	encoded = append(encoded, bytes...)

	return encoded, nil
}

func (utfStringPair *UtfStringPair) Decode(input []byte) error {
	// TODO: Decode utf string pair
	return nil
}

// Encodes binary data as:
// | size MSB | size LSB | binary data |
func (binaryData *BinaryData) Encode() ([]byte, error) {
	encoded := make([]byte, 0, len(binaryData.Data)+2)

	length := UnsignedInt{Value: uint32(len(binaryData.Data)), Size: 2}
	encLength, err := length.Encode()
	if err != nil {
		return []byte{},
			errors.Join(codec.EncodeErr(binaryData, "length encoding error"), err)
	}

	copy(encoded, encLength)
	copy(encoded[length.Value:], binaryData.Data)

	return encoded, nil
}

func (vbi *VariableByteInteger) Encode() (bin []byte, err error) {
	x := vbi.Value

	for {
		encodedByte := x % 128
		x = x / 128
		if x > 0 {
			encodedByte |= 128
		}
		bin = append(bin, byte(encodedByte))
		if x <= 0 {
			break
		}
	}

	return bin, nil
}

func (binaryData *BinaryData) Decode(input []byte) (int, error) {
  length := UnsignedInt{Size: 2}
  n, err := length.Decode(input)
  if err != nil {
    return 0, err
  }
  fmt.Println("payload length", length.Value)

  input = input[n:]

  binaryData.Data = input[:length.Value]

  return n+int(length.Value), nil
}

func (vbi *VariableByteInteger) Decode(input []byte) (int, error) {
	if len(input) < 1 {
		return 0, codec.DecodeErr(vbi, "input length < 1")
	}

	multiplier := 1
	value := 0
	i := 0
	fmt.Println("decoding vbi", input[i], input[i]&128)
	for i = 0; i < 4; i++ {
		fmt.Println(input[i], value, multiplier)
		value += int(input[i]&127) * multiplier
		if multiplier > 128*128*128 {
			return 0, codec.DecodeErr(vbi, "malformed variable byte integer")
		}
		multiplier *= 128

		if (input[i] & 128) == 0 {
			break
		}
	}
	vbi.Value = int32(value)

	return i + 1, nil
}
