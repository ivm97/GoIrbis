package irbis

import (
	"strconv"
)

const defaultQueryCap = 256

// ClientQuery builds one IRBIS client request body.
type ClientQuery struct {
	buf []byte
}

// NewClientQuery builds a request header and advances QueryId for the session.
func NewClientQuery(connection *Connection, command string) *ClientQuery {
	query := &ClientQuery{buf: make([]byte, 0, defaultQueryCap)}
	query.AddAnsi(command).NewLine()
	query.AddAnsi(connection.Workstation).NewLine()
	query.AddAnsi(command).NewLine()
	query.Add(connection.ClientId).NewLine()
	query.Add(connection.QueryId).NewLine()
	query.AddAnsi(connection.Password).NewLine()
	query.AddAnsi(connection.Username).NewLine()
	query.NewLine()
	query.NewLine()
	query.NewLine()
	connection.QueryId++
	return query
}

// Add appends an integer in ANSI/UTF-safe decimal form.
func (query *ClientQuery) Add(value int) *ClientQuery {
	query.buf = strconv.AppendInt(query.buf, int64(value), 10)
	return query
}

// AddAnsi appends text encoded as Windows-1251.
func (query *ClientQuery) AddAnsi(text string) *ClientQuery {
	query.buf = appendCP1251(query.buf, text)
	return query
}

// AddFormat appends a prepared format line and a newline.
func (query *ClientQuery) AddFormat(format string) bool {
	if len(format) == 0 {
		query.NewLine()
		return false
	}

	prepared := prepareFormat(trimLeft(format))

	if format[0] == '@' {
		query.AddAnsi(prepared)
	} else if format[0] == '!' {
		query.AddUtf(prepared)
	} else {
		query.AddUtf("!")
		query.AddUtf(prepared)
	}
	query.NewLine()
	return true
}

// AddUtf appends text as UTF-8 bytes.
func (query *ClientQuery) AddUtf(text string) *ClientQuery {
	query.buf = append(query.buf, text...)
	return query
}

// Encode returns the packet as a single chunk (compatibility helper).
func (query *ClientQuery) Encode() [][]byte {
	return [][]byte{query.EncodePacket()}
}

// EncodePacket builds length-prefix + body for one TCP write.
func (query *ClientQuery) EncodePacket() []byte {
	bodyLen := len(query.buf)
	prefix := strconv.AppendInt(make([]byte, 0, 16), int64(bodyLen), 10)
	prefix = append(prefix, '\n')

	packet := make([]byte, 0, len(prefix)+bodyLen)
	packet = append(packet, prefix...)
	packet = append(packet, query.buf...)
	return packet
}

// NewLine appends '\n'.
func (query *ClientQuery) NewLine() *ClientQuery {
	query.buf = append(query.buf, '\n')
	return query
}
