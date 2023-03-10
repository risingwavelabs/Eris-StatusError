# Eris with status codes ![Logo][eris-logo]

[![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![GoReport][report-img]][report] [![Coverage Status][cov-img]][cov]

Package `eris with status codes` is an error handling library with readable stack traces, flexible formatting support, error codes and custom key-value pairs.

`go get github.com/risingwavelabs/eris`

<!-- toc -->

- [Eris with status codes ](#eris-with-status-codes-)
  - [Why you should switch from eris to eris with status codes](#why-you-should-switch-from-eris-to-eris-with-status-codes)
  - [Status codes](#status-codes)
  - [Using eris with status codes](#using-eris-with-status-codes)
    - [Creating errors](#creating-errors)
    - [Wrapping errors](#wrapping-errors)
  - [Handling errors](#handling-errors)

<!-- tocstop -->

## Why you should switch from eris to eris with status codes

Eris with status codes is build on top of [rotisserie/eris](https://github.com/rotisserie/eris). A error handling library that enables you to write errors in JSON form and to wrap errors easily.  

We extended the error library to include a status code and additional custom parameters that you can set during error handling. We aim for `eris` to be compatible with `rotisserie/eris`. At the time of this writing you can use our `eris` as a drop-in replacement fro `rotisserie/eris`.

This document will focus on the extension of `rotisserie/eris`. Please have a look at the [rotisserie/eris README](https://github.com/rotisserie/eris) for information about basic functionalities of this library, or just view our [pkg documentation](https://pkg.go.dev/github.com/risingwavelabs/eris@v0.0.0-20230309150549-b35f50e98d1a#section-readme)

```json
{
    "external": "external error",
    "root": {
        "KVs": {
            "bar": 42,
            "foo": true
        },
        "code": "data loss",
        "message": "no good",
        "stack": [
            "eris_test.TestMain:/path/to/file.go:370",
            "eris_test.TestMain:/path/to/other.go:370",
        ]
    },
    "wrap": [
        {
            "code": "internal",
            "message": "even more context",
            "stack": "eris_test.TestMain:/path/to/another.go:370",
        }
    ]
}
```

## Status codes

We are using [GRPC status codes](https://grpc.github.io/grpc/core/md_doc_statuscodes.html)

| Status Code      | Description |
| ----------- | ----------- |
| Canceled | The operation was cancelled, typically by the caller. |
| Unknown | Unknown error. For example, this error may be returned when a Status value received from another address space belongs to an error space that is not known in this address space. Also errors raised by APIs that do not return enough error information may be converted to this error. |
| InvalidArgument | The client specified an invalid argument. Note that this differs from FailedPrecondition. InvalidArgument indicates arguments that are problematic regardless of the state of the system (e.g., a malformed file name). |
| DeadlineExceeded | The deadline expired before the operation could complete. For operations that change the state of the system, this error may be returned even if the operation has completed successfully. For example, a successful response from a server could have been delayed long. |
| NotFound | Some requested entity (e.g., file or directory) was not found. Note to server developers: if a request is denied for an entire class of users, such as gradual feature rollout or undocumented allowlist, NotFound may be used. If a request is denied for some users within a class of users, such as user-based access control, PermissionDenied must be used. |
| AlreadyExists | The entity that a client attempted to create (e.g., file or directory) already exists. |
| PermissionDenied | The caller does not have permission to execute the specified operation. PermissionDenied must not be used for rejections caused by exhausting some resource (use ResourceExhausted instead for those errors). PermissionDenied must not be used if the caller can not be identified (use Unauthenticated instead for those errors). This error code does not imply the request is valid or the requested entity exists or satisfies other pre-conditions. |
| ResourceExhausted | Some resource has been exhausted, perhaps a per-user quota, or perhaps the entire file system is out of space. |
| FailedPrecondition | The operation was rejected because the system is not in a state required for the operation's execution. For example, the directory to be deleted is non-empty, an rmdir operation is applied to a non-directory, etc. Service implementors can use the following guidelines to decide between FailedPrecondition, Aborted, and unavailable: (a) Use unavailable if the client can retry just the failing call. (b) Use Aborted if the client should retry at a higher level (e.g., when a client-specified test-and-set fails, indicating the client should restart a read-modify-write sequence). (c) Use FailedPrecondition if the client should not retry until the system state has been explicitly fixed. E.g., if an "rmdir" fails because the directory is non-empty, FailedPrecondition should be returned since the client should not retry unless the files are deleted from the directory. |
| Aborted | The operation was aborted, typically due to a concurrency issue such as a sequencer check failure or transaction abort. See the guidelines above for deciding between FailedPrecondition, Aborted, and unavailable. |
| OutOfRange | The operation was attempted past the valid range. E.g., seeking or reading past end-of-file. Unlike InvalidArgument, this error indicates a problem that may be fixed if the system state changes. For example, a 32-bit file system will generate InvalidArgument if asked to read at an offset that is not in the range [0,2^32-1], but it will generate OutOfRange if asked to read from an offset past the current file size. There is a fair bit of overlap between FailedPrecondition and OutOfRange. We recommend using OutOfRange (the more specific error) when it applies so that callers who are iterating through a space can easily look for an OutOfRange error to detect when they are done. |
| Unimplemented | The operation is not implemented or is not supported/enabled in this service. |
| Internal | Internal errors. This means that some invariants expected by the underlying system have been broken. This error code is reserved for serious errors. |
| Unavailable | The service is currently unavailable. This is most likely a transient condition, which can be corrected by retrying with a back off. Note that it is not always safe to retry non-idempotent operations. |
| DataLoss | Unrecoverable data loss or corruption. |
| Unauthenticated | The request does not have valid authentication credentials for the operation. |

**Defaults**: The status code `Unknown` will be used by default when calling `eris.New` and the code `Internal` when calling `eris.Wrap`.

You can convert between `eris.Code` and `grpc.Code` using `FromGrpc(c grpc.Code) Code` and `(c Code) ToGRPC() grpc.Code`.

## Using eris with status codes

### Creating errors


Creating errors is simple via [`eris.New`](https://pkg.go.dev/github.com/risingwavelabs/eris#New). The default assigned error code will be `unknown`. If you would like to assign a different error, use `WithCode`. You can also assign additional properties using `WithProperty`.

```golang
func (req *Request) Validate() error {
  if req.ID == "" {
    return eris.New("error bad request")
      .WithCode(eris.CodeInvalidArgument)
      .WithProperty("request ID", "")
  }
  return nil
}
```

### Wrapping errors

[`eris.Wrap`](https://pkg.go.dev/github.com/risingwavelabs/eris#Wrap) adds context to an error while preserving the original error. The default assigned error code will be `internal`. Like above you can change the code via `WithCode` and set additional properties using `WithProperty`.

```golang
relPath, err := GetRelPath("/Users/roti/", resource.AbsPath)
if err != nil {
  // wrap the error if you want to add more context
  return nil, eris.Wrapf(err, "failed to get relative path for resource '%v'", resource.ID) // has code internal
}
```

## Handling errors 

When handling errors, you can rely on the error code. You may also use the additional properties to get more information about the cause of the error. 

```go 
err := operation()

if err.Code() == eris.CodeDeadlineExceeded {
	// retry operation
}

if err.HasKVs() {
	additional_context := caller2.KVs()
	// log additional context
}
```

You can also use `GetCode(err error)`. This will default to `unknown` if you pass in an standard lib error. 



-----------------------------------------------------------------

Released under the [MIT License].

[MIT License]: LICENSE.txt
[eris-logo]: https://cdn.emojidex.com/emoji/hdpi/minecraft_golden_apple.png?1511637499
[doc-img]: https://pkg.go.dev/badge/github.com/risingwavelabs/eris
[doc]: https://pkg.go.dev/github.com/risingwavelabs/eris
[ci-img]: https://github.com/risingwavelabs/eris/workflows/build/badge.svg
[ci]: https://github.com/risingwavelabs/eris/actions
[report-img]: https://goreportcard.com/badge/github.com/risingwavelabs/eris
[report]: https://goreportcard.com/report/github.com/risingwavelabs/eris
[cov-img]: https://codecov.io/gh/risingwavelabs/eris/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/risingwavelabs/eris

