package eris

// Code is an error code that indicates the category of error.
type Code int

// These are common impact error codes that are found throughout our services.
const (
	// TODO: Change these to GRPC codes without ok
	// https://grpc.github.io/grpc/core/md_doc_statuscodes.html

	// CodeUnknown is the default code for errors that are not classified.
	CodeUnknown Code = iota
	// CodeAlreadyExists means an attempt to create an entity failed because one.
	// already exists.
	CodeAlreadyExists
	// CodeNotFound means some requested entity (e.g., file or directory) was not found.
	CodeNotFound
	// CodeInvalidArgument indicates that the caller specified an invalid argument.
	CodeInvalidArgument
	// CodeMalformedRequest indicates the syntax of the request cannot be interpreted (eg JSON decoding error).
	CodeMalformedRequest
	// CodeUnauthenticated indicates the request does not have valid authentication credentials for the operation.
	CodeUnauthenticated
	// CodePermissionDenied indicates that the identity of the user is confirmed but they do not have permissions.
	// to perform the request.
	CodePermissionDenied
	// CodeConstraintViolated indicates that a constraint in the system has been violated.
	// Eg. a duplicate key error from a unique index.
	CodeConstraintViolated
	// CodeNotSupported indicates that the request is not supported.
	CodeNotSupported
	// CodeNotImplemented indicates that the request is not implemented.
	CodeNotImplemented
	// CodeMissingParameter indicates that a required parameter is missing or empty.
	CodeMissingParameter
	// CodeDeadlineExceeded indicates that a request exceeded it's deadline before completion.
	CodeDeadlineExceeded
	// CodeCanceled indicates that the request was canceled before completion.
	CodeCanceled
	// CodeResourceExhausted indicates that some limited resource (eg rate limit or disk space) has been reached.
	CodeResourceExhausted
	// CodeUnavailable indicates that the server itself is unavailable for processing requests.
	CodeUnavailable
	// Default error for wrap.
	CodeInternal
)

// TODO: rename this to grpc error codes.
func (c Code) String() string {
	if s, ok := defaultErrorCodes[c]; ok {
		return s
	}
	return defaultErrorCodes[CodeUnknown]
}

var defaultErrorCodes = map[Code]string{
	CodeUnknown:            "unknown",
	CodeAlreadyExists:      "already exists",
	CodeNotFound:           "not found",
	CodeInvalidArgument:    "invalid argument",
	CodeMalformedRequest:   "malformed request",
	CodeUnauthenticated:    "unauthenticated",
	CodePermissionDenied:   "permission denied",
	CodeConstraintViolated: "constraint violated",
	CodeNotSupported:       "not supported",
	CodeMissingParameter:   "parameter is missing",
	CodeNotImplemented:     "not implemented",
	CodeDeadlineExceeded:   "deadline exceeded",
	CodeCanceled:           "canceled",
	CodeResourceExhausted:  "resource exhausted",
	CodeUnavailable:        "unavailable",
	CodeInternal:           "internal",
}

const (
	// Default error code assigned when using eris.New.
	DEFAULT_ERROR_CODE_NEW = CodeUnknown
	// Default error code assigned when using eris.Wrap or Wrapf.
	DEFAULT_ERROR_CODE_WRAP = CodeInternal
)

// TODO Functions should use default code in default cases and never actual codes directly.
