package irbis

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const IsoMarkerLength = 24

const (
	IsoRecordDelimiter   = byte(0x1D)
	IsoFieldDelimiter    = byte(0x1E)
	IsoSubfieldDelimiter = byte(0x1F)
)

// ErrInvalidISO means the stream is not a well-formed ISO2709 record.
var ErrInvalidISO = errors.New("not an ISO2709 record")

func encodeInt32(buffer []byte, position, length, value int) {
	length--
	for position += length; length >= 0; length-- {
		buffer[position] = byte(value%10) + byte('0')
		value /= 10
		position--
	}
}

func encodeText(buffer []byte, position int, text string) int {
	if len(text) != 0 {
		encoded := []byte(text)
		for i := 0; i < len(encoded); i++ {
			buffer[position] = encoded[i]
			position++
		}
	}

	return position
}

// DecodeBody decodes only the field body text.
func (field *RecordField) decodeBody(body string) {
	delimiter := IsoSubfieldDelimiter
	all := strings.Split(body, string(delimiter))
	if body[0] != delimiter {
		field.Value = all[0]
		all = all[1:]
	}
	for _, one := range all {
		if len(one) != 0 {
			subfield := new(SubField)
			subfield.Decode(one)
			field.Subfields = append(field.Subfields, subfield)
		}
	}
}

// ReadIsoRecord reads one ISO2709 record from reader.
func ReadIsoRecord(reader io.Reader, decoder func([]byte) string) (*MarcRecord, error) {
	result := NewMarcRecord()

	marker := make([]byte, 5)
	if _, err := io.ReadFull(reader, marker); err != nil {
		return nil, err
	}

	recordLength := ParseInt32(marker)
	if recordLength < IsoMarkerLength {
		return nil, fmt.Errorf("%w: invalid length %d", ErrInvalidISO, recordLength)
	}

	record := make([]byte, recordLength)
	copy(record, marker)
	if _, err := io.ReadFull(reader, record[len(marker):]); err != nil {
		return nil, err
	}

	if record[recordLength-1] != IsoRecordDelimiter {
		return nil, ErrInvalidISO
	}

	lengthOfLength := ParseInt32(record[20:21])
	lengthOfOffset := ParseInt32(record[21:22])
	additionalData := ParseInt32(record[22:23])
	directoryLength := 3 + lengthOfLength + lengthOfOffset + additionalData
	indicatorLength := ParseInt32(record[10:11])
	baseAddress := ParseInt32(record[12:17])

	fieldCount := 0
	for ofs := IsoMarkerLength; ; ofs += directoryLength {
		if record[ofs] == IsoFieldDelimiter {
			break
		}
		fieldCount++
	}
	result.Fields = make([]*RecordField, 0, fieldCount)

	for directory := IsoMarkerLength; ; directory += directoryLength {
		if record[directory] == IsoFieldDelimiter {
			break
		}

		tag := ParseInt32(record[directory : directory+3])
		ofs := directory + 3
		fieldLength := ParseInt32(record[ofs : ofs+lengthOfLength])
		ofs = directory + 3 + lengthOfLength
		fieldOffset := baseAddress + ParseInt32(record[ofs:ofs+lengthOfOffset])
		field := NewRecordField(tag, "")
		result.Fields = append(result.Fields, field)
		if tag < 10 {
			temp := record[fieldOffset : fieldOffset+fieldLength-1]
			field.Value = decoder(temp)
		} else {
			start := fieldOffset + indicatorLength
			stop := fieldOffset + fieldLength - indicatorLength + 1
			temp := record[start:stop]
			text := decoder(temp)
			field.decodeBody(text)
		}
	}

	return result, nil
}
