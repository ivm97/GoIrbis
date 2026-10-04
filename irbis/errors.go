package irbis

import (
	"errors"
	"fmt"
)

// IRBIS64 and client error codes (negative values).
const (
	CodeMFNOutOfBounds         = -100
	CodeInvalidShelfSize       = -101
	CodeInvalidShelfNumber     = -102
	CodeMFNOutOfBoundsAlt      = -140
	CodeReadError              = -141
	CodeFieldMissing           = -200
	CodePrevVersionMissing     = -201
	CodeTermNotFound           = -202
	CodeLastTerm               = -203
	CodeFirstTerm              = -204
	CodeDBExclusiveLock        = -300
	CodeDBExclusiveLockAlt     = -301
	CodeMSTOpenError           = -400
	CodeIFPOpenError           = -401
	CodeWriteError             = -402
	CodeActualizeError         = -403
	CodeRecordLogicallyDeleted = -600
	CodeRecordPhysicallyDeleted = -601
	CodeRecordLocked           = -602
	CodeRecordLogicallyDeletedAlt = -603
	CodeRecordPhysicallyDeletedAlt = -605
	CodeAutoinError            = -607
	CodeRecordVersionConflict  = -608
	CodeBackupCreateError      = -700
	CodeBackupRestoreError     = -701
	CodeSortError              = -702
	CodeInvalidTerm            = -703
	CodeDictionaryCreateError  = -704
	CodeDictionaryLoadError    = -705
	CodeGBLInvalidParams       = -800
	CodeGBLRep                 = -801
	CodeGBLMet                 = -802
	CodeServerExecuteError     = -1111
	CodeWrongProtocol          = -2222
	CodeUnregisteredClient     = -3333
	CodeClientNotLoggedIn      = -3334
	CodeInvalidClientID        = -3335
	CodeNoWorkstationAccess    = -3336
	CodeClientAlreadyRegistered = -3337
	CodeClientNotAllowed       = -3338
	CodeWrongPassword          = -4444
	CodeFileNotFound           = -5555
	CodeServerOverloaded       = -6666
	CodeAdminProcessError      = -7777
	CodeGeneralError           = -8888

	// CodeNetwork is a client-side transport failure (not an IRBIS server code).
	CodeNetwork = -100000
)

// ErrCodeNetwork is an alias kept for older call sites.
const ErrCodeNetwork = CodeNetwork

// Sentinel errors for errors.Is checks. Values are immutable templates;
// NewError/WrapError always return a fresh *Error instance.
var (
	ErrMFNOutOfBounds          = newSentinel(CodeMFNOutOfBounds)
	ErrInvalidShelfSize        = newSentinel(CodeInvalidShelfSize)
	ErrInvalidShelfNumber      = newSentinel(CodeInvalidShelfNumber)
	ErrReadError               = newSentinel(CodeReadError)
	ErrFieldMissing            = newSentinel(CodeFieldMissing)
	ErrPrevVersionMissing      = newSentinel(CodePrevVersionMissing)
	ErrTermNotFound            = newSentinel(CodeTermNotFound)
	ErrLastTerm                = newSentinel(CodeLastTerm)
	ErrFirstTerm               = newSentinel(CodeFirstTerm)
	ErrDBExclusiveLock         = newSentinel(CodeDBExclusiveLock)
	ErrMSTOpenError            = newSentinel(CodeMSTOpenError)
	ErrIFPOpenError            = newSentinel(CodeIFPOpenError)
	ErrWriteError              = newSentinel(CodeWriteError)
	ErrActualizeError          = newSentinel(CodeActualizeError)
	ErrRecordLogicallyDeleted  = newSentinel(CodeRecordLogicallyDeleted)
	ErrRecordPhysicallyDeleted = newSentinel(CodeRecordPhysicallyDeleted)
	ErrRecordLocked            = newSentinel(CodeRecordLocked)
	ErrAutoinError             = newSentinel(CodeAutoinError)
	ErrRecordVersionConflict   = newSentinel(CodeRecordVersionConflict)
	ErrBackupCreateError       = newSentinel(CodeBackupCreateError)
	ErrBackupRestoreError      = newSentinel(CodeBackupRestoreError)
	ErrSortError               = newSentinel(CodeSortError)
	ErrInvalidTerm             = newSentinel(CodeInvalidTerm)
	ErrDictionaryCreateError   = newSentinel(CodeDictionaryCreateError)
	ErrDictionaryLoadError     = newSentinel(CodeDictionaryLoadError)
	ErrGBLInvalidParams        = newSentinel(CodeGBLInvalidParams)
	ErrGBLRep                  = newSentinel(CodeGBLRep)
	ErrGBLMet                  = newSentinel(CodeGBLMet)
	ErrServerExecuteError      = newSentinel(CodeServerExecuteError)
	ErrWrongProtocol           = newSentinel(CodeWrongProtocol)
	ErrUnregisteredClient      = newSentinel(CodeUnregisteredClient)
	ErrClientNotLoggedIn       = newSentinel(CodeClientNotLoggedIn)
	ErrInvalidClientID         = newSentinel(CodeInvalidClientID)
	ErrNoWorkstationAccess     = newSentinel(CodeNoWorkstationAccess)
	ErrClientAlreadyRegistered = newSentinel(CodeClientAlreadyRegistered)
	ErrClientNotAllowed        = newSentinel(CodeClientNotAllowed)
	ErrWrongPassword           = newSentinel(CodeWrongPassword)
	ErrFileNotFound            = newSentinel(CodeFileNotFound)
	ErrServerOverloaded        = newSentinel(CodeServerOverloaded)
	ErrAdminProcessError       = newSentinel(CodeAdminProcessError)
	ErrGeneralError            = newSentinel(CodeGeneralError)
	ErrNetwork                 = newSentinel(CodeNetwork)
)

func newSentinel(code int) *Error {
	return &Error{Code: code, Message: describeCode(code)}
}

// Error is a standardized IRBIS/client failure.
type Error struct {
	Code    int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	msg := e.Message
	if msg == "" {
		msg = DescribeError(e.Code)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is reports whether target is an *Error with the same Code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || e == nil || t == nil {
		return false
	}
	return e.Code == t.Code
}

// NewError builds a fresh error from an IRBIS/client code.
func NewError(code int) *Error {
	return &Error{
		Code:    code,
		Message: DescribeError(code),
	}
}

// WrapError attaches a cause to an IRBIS/client code.
func WrapError(code int, err error) *Error {
	return &Error{
		Code:    code,
		Message: DescribeError(code),
		Err:     err,
	}
}

// ReturnCodeError returns nil for non-negative codes, otherwise NewError(code).
func ReturnCodeError(code int) error {
	if code >= 0 {
		return nil
	}
	return NewError(code)
}

// CodeOf returns the IRBIS/client code, or 0 if err is not *Error.
func CodeOf(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 0
}

// AsError extracts *Error from err.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// DescribeError returns an English description for an IRBIS/client return code.
func DescribeError(code int) string {
	if code >= 0 {
		return "no error"
	}
	if msg := describeCode(code); msg != "" {
		return msg
	}
	return fmt.Sprintf("unknown error (%d)", code)
}

func describeCode(code int) string {
	switch code {
	case CodeMFNOutOfBounds, CodeMFNOutOfBoundsAlt:
		return "MFN is outside the database bounds"
	case CodeInvalidShelfSize:
		return "invalid shelf size"
	case CodeInvalidShelfNumber:
		return "invalid shelf number"
	case CodeReadError:
		return "read error"
	case CodeFieldMissing:
		return "requested field is missing"
	case CodePrevVersionMissing:
		return "previous record version is missing"
	case CodeTermNotFound:
		return "term not found"
	case CodeLastTerm:
		return "last term in the list"
	case CodeFirstTerm:
		return "first term in the list"
	case CodeDBExclusiveLock, CodeDBExclusiveLockAlt:
		return "database is exclusively locked"
	case CodeMSTOpenError:
		return "failed to open MST or XRF (master data file error)"
	case CodeIFPOpenError:
		return "failed to open IFP (index file error)"
	case CodeWriteError:
		return "write error"
	case CodeActualizeError:
		return "actualization error"
	case CodeRecordLogicallyDeleted, CodeRecordLogicallyDeletedAlt:
		return "record is logically deleted"
	case CodeRecordPhysicallyDeleted, CodeRecordPhysicallyDeletedAlt:
		return "record is physically deleted"
	case CodeRecordLocked:
		return "record is locked for edit"
	case CodeAutoinError:
		return "autoin.gbl error"
	case CodeRecordVersionConflict:
		return "record version conflict"
	case CodeBackupCreateError:
		return "backup creation failed"
	case CodeBackupRestoreError:
		return "restore from backup failed"
	case CodeSortError:
		return "sort error"
	case CodeInvalidTerm:
		return "invalid term"
	case CodeDictionaryCreateError:
		return "dictionary creation failed"
	case CodeDictionaryLoadError:
		return "dictionary load failed"
	case CodeGBLInvalidParams:
		return "invalid global correction parameters"
	case CodeGBLRep:
		return "global correction: ERR_GBL_REP"
	case CodeGBLMet:
		return "global correction: ERR_GBL_MET"
	case CodeServerExecuteError:
		return "server execution error"
	case CodeWrongProtocol:
		return "protocol error"
	case CodeUnregisteredClient:
		return "unregistered client (not in the client list)"
	case CodeClientNotLoggedIn:
		return "client has not logged in"
	case CodeInvalidClientID:
		return "invalid client identifier"
	case CodeNoWorkstationAccess:
		return "no access to workstation commands"
	case CodeClientAlreadyRegistered:
		return "client is already registered"
	case CodeClientNotAllowed:
		return "client is not allowed"
	case CodeWrongPassword:
		return "wrong password"
	case CodeFileNotFound:
		return "file does not exist"
	case CodeServerOverloaded:
		return "server overloaded (max processing threads reached)"
	case CodeAdminProcessError:
		return "failed to start or stop administrator process"
	case CodeGeneralError:
		return "general error"
	case CodeNetwork:
		return "network error: failed to connect to server"
	default:
		return ""
	}
}
