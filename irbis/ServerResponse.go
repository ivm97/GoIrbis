package irbis

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"strings"
)

// ServerResponse is a parsed IRBIS server reply.
type ServerResponse struct {
	Command       string
	ClientId      int
	QueryId       int
	AnswerSize    int
	ReturnCode    int
	ServerVersion string
	reader        *bytes.Reader
	connection    *Connection
}

// ReadServerResponse reads and parses a full response from conn.
func ReadServerResponse(conn net.Conn) (*ServerResponse, error) {
	buffer, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return ParseServerResponse(buffer), nil
}

// ParseServerResponse parses an already buffered server reply.
func ParseServerResponse(buffer []byte) *ServerResponse {
	result := &ServerResponse{
		reader: bytes.NewReader(buffer),
	}
	result.Command = result.ReadAnsi()
	result.ClientId = result.ReadInteger()
	result.QueryId = result.ReadInteger()
	result.AnswerSize = result.ReadInteger()
	result.ServerVersion = result.ReadAnsi()
	result.ReadAnsi()
	result.ReadAnsi()
	result.ReadAnsi()
	result.ReadAnsi()
	result.ReadAnsi()
	return result
}

// NewServerResponse reads a response from conn.
// Deprecated: prefer ReadServerResponse.
func NewServerResponse(conn net.Conn) *ServerResponse {
	response, err := ReadServerResponse(conn)
	if err != nil {
		return nil
	}
	return response
}

func (response *ServerResponse) CheckReturnCode(allowed ...int) bool {
	if response.GetReturnCode() < 0 {
		if contains(allowed, response.ReturnCode) {
			return true
		}
		return false
	}

	return true
}

func (response *ServerResponse) GetLine() []byte {
	result := bytes.Buffer{}
	for response.reader.Len() != 0 {
		one, err := response.reader.ReadByte()
		if err != nil {
			break
		}
		if one == 13 {
			one, err = response.reader.ReadByte()
			if err != nil {
				break
			}
			if one != 10 {
				_ = response.reader.UnreadByte()
			}
			break
		}

		result.WriteByte(one)
	}

	return result.Bytes()
}

func (response *ServerResponse) GetReturnCode() int {
	response.ReturnCode = response.ReadInteger()
	if response.connection != nil {
		response.connection.LastError = response.ReturnCode
		if response.ReturnCode < 0 {
			response.connection.setError(NewError(response.ReturnCode))
		} else {
			response.connection.clearError()
		}
	}
	return response.ReturnCode
}

func (response *ServerResponse) ReadAnsi() string {
	line := response.GetLine()
	result := FromAnsi(line)
	return result
}

func (response *ServerResponse) ReadInteger() int {
	result, _ := strconv.Atoi(response.ReadAnsi())
	return result
}

func (response *ServerResponse) ReadRemainingAnsiLines() []string {
	text := response.ReadRemainingAnsiText()
	result := strings.Split(text, "\n")
	for i := range result {
		result[i] = strings.ReplaceAll(result[i], "\r", "")
	}
	return result
}

func (response *ServerResponse) ReadRemainingAnsiText() string {
	line, _ := io.ReadAll(response.reader)
	result := FromAnsi(line)

	return result
}

func (response *ServerResponse) ReadRemainingUtfLines() []string {
	text := response.ReadRemainingUtfText()
	result := strings.Split(text, "\n")
	for i := range result {
		result[i] = strings.TrimSuffix(result[i], "\r")
	}
	return result
}

func (response *ServerResponse) ReadRemainingUtfText() string {
	line, _ := io.ReadAll(response.reader)
	result := fromUtf8(line)

	return result
}

func (response *ServerResponse) ReadUtf() string {
	line := response.GetLine()
	result := fromUtf8(line)
	return result
}
