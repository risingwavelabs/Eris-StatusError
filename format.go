package eris

import (
	"fmt"
)

// FormatOptions defines output options like omitting stack traces and inverting the error or stack order.
type FormatOptions struct {
	InvertOutput bool // Flag that inverts the error output (wrap errors shown first).
	WithTrace    bool // Flag that enables stack trace output.
	InvertTrace  bool // Flag that inverts the stack trace output (top of call stack shown first).
	WithExternal bool // Flag that enables external error output.
	// todo: maybe allow users to hide wrap frames if desired
}

// StringFormat defines a string error format.
type StringFormat struct {
	Options      FormatOptions // Format options (e.g. omitting stack trace or inverting the output order).
	MsgStackSep  string        // Separator between error messages and stack frame data.
	PreStackSep  string        // Separator at the beginning of each stack frame.
	StackElemSep string        // Separator between elements of each stack frame.
	ErrorSep     string        // Separator between each error in the chain.
}

// NewDefaultStringFormat returns a default string output format.
func NewDefaultStringFormat(options FormatOptions) StringFormat {
	stringFmt := StringFormat{
		Options: options,
	}
	if options.WithTrace {
		stringFmt.MsgStackSep = "\n"
		stringFmt.PreStackSep = "\t"
		stringFmt.StackElemSep = ":"
		stringFmt.ErrorSep = "\n"
	} else {
		stringFmt.ErrorSep = ": "
	}
	return stringFmt
}

// ToString returns a default formatted string for a given error.
//
// An error without trace will be formatted as follows:
//
//	<Wrap error msg>: <Root error msg>
//
// An error with trace will be formatted as follows:
//
//	<Wrap error msg>
//	  <Method2>:<File2>:<Line2>
//	<Root error msg>
//	  <Method2>:<File2>:<Line2>
//	  <Method1>:<File1>:<Line1>
func ToString(err error, withTrace bool) string {
	return ToCustomString(err, NewDefaultStringFormat(FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomString returns a custom formatted string for a given error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as follows:
//
//	<Wrap error msg>[Format.ErrorSep]<Root error msg>
//
// An error with trace will be formatted as follows:
//
//	<Wrap error msg>[Format.MsgStackSep]
//	[Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
//	<Root error msg>[Format.MsgStackSep]
//	[Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
//	[Format.PreStackSep]<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>[Format.ErrorSep]
func ToCustomString(err error, format StringFormat) string {
	upErr := Unpack(err)

	var str string
	if format.Options.InvertOutput {
		if format.Options.WithExternal && upErr.ErrExternal != nil {
			str += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Msg != "" {
				str += format.ErrorSep
			}
		}
		str += upErr.ErrRoot.formatStr(format)
		for _, eLink := range upErr.ErrChain {
			str += format.ErrorSep + eLink.formatStr(format)
		}
	} else {
		for i := len(upErr.ErrChain) - 1; i >= 0; i-- {
			str += upErr.ErrChain[i].formatStr(format) + format.ErrorSep
		}
		str += upErr.ErrRoot.formatStr(format)
		if format.Options.WithExternal && upErr.ErrExternal != nil {
			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Msg != "" {
				str += format.ErrorSep
			}
			str += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
		}
	}

	return str
}

// JSONFormat defines a JSON error format.
type JSONFormat struct {
	Options FormatOptions // Format options (e.g. omitting stack trace or inverting the output order).
	// todo: maybe allow setting of wrap/root keys in the output map as well
	StackElemSep string // Separator between elements of each stack frame.
}

// NewDefaultJSONFormat returns a default JSON output format.
func NewDefaultJSONFormat(options FormatOptions) JSONFormat {
	return JSONFormat{
		Options:      options,
		StackElemSep: ":",
	}
}

// ToJSON returns a JSON formatted map for a given error.
//
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	      "message": "Root error msg"
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg"
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>:<File2>:<Line2>",
//	      "<Method1>:<File1>:<Line1>"
//	    ]
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>:<File2>:<Line2>"
//	    }
//	  ]
//	}
func ToJSON(err error, withTrace bool) map[string]any {
	return ToCustomJSON(err, NewDefaultJSONFormat(FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// TODO: change this documentation comment. KVs and code
// ToCustomJSON returns a JSON formatted map for a given error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg'",
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>",
//	      "<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>"
//	    ]
//	  }
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>"
//	    }
//	  ]
//	}
func ToCustomJSON(err error, format JSONFormat) map[string]any {
	upErr := Unpack(err)

	jsonMap := make(map[string]any)
	if format.Options.WithExternal && upErr.ErrExternal != nil {
		jsonMap["external"] = formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
	}

	if upErr.ErrRoot.Msg != "" || len(upErr.ErrRoot.Stack) > 0 {
		jsonMap["root"] = upErr.ErrRoot.formatJSON(format)
	}
	if len(upErr.ErrChain) > 0 {
		var wrapArr []map[string]any
		for _, eLink := range upErr.ErrChain {
			wrapMap := eLink.formatJSON(format)
			if format.Options.InvertOutput {
				wrapArr = append(wrapArr, wrapMap)
			} else {
				wrapArr = append([]map[string]any{wrapMap}, wrapArr...)
			}
		}
		jsonMap["wrap"] = wrapArr
	}

	return jsonMap
}

// Unpack returns a human-readable UnpackedError type for a given error.
func Unpack(err error) UnpackedError {
	var upErr UnpackedError
	for err != nil {
		switch err := err.(type) {
		case *rootError:
			upErr.ErrRoot.Msg = err.msg
			upErr.ErrRoot.Stack = err.stack.get()
			upErr.ErrRoot.code = err.code
			upErr.ErrRoot.KVs = err.KVs
		case *wrapError:
			// prepend links in stack trace order
			link := ErrLink{Msg: err.msg}
			link.Frame = err.frame.get()
			link.code = err.code
			link.KVs = err.KVs
			upErr.ErrChain = append([]ErrLink{link}, upErr.ErrChain...)
		default:
			upErr.ErrExternal = err
			return upErr
		}
		err = Unwrap(err)
	}
	return upErr
}

// UnpackedError represents complete information about an error.
//
// This type can be used for custom error logging and parsing. Use `eris.Unpack` to build an UnpackedError
// from any error type. The ErrChain and ErrRoot fields correspond to `wrapError` and `rootError` types,
// respectively. If any other error type is unpacked, it will appear in the ExternalErr field.
type UnpackedError struct {
	ErrExternal error
	ErrRoot     ErrRoot
	ErrChain    []ErrLink
}

// String formatter for external errors.
func formatExternalStr(err error, withTrace bool) string {
	if withTrace {
		return fmt.Sprintf("%+v", err)
	}
	return fmt.Sprint(err)
}

// ErrRoot represents an error stack and the accompanying message.
type ErrRoot struct {
	Msg   string
	Stack Stack
	code  Code
	KVs   map[string]any // TODO: do not expose kvs field in the different error types?
}

// HasKVs returns true if the error has key-value pairs.
func (err *ErrRoot) HasKVs() bool {
	return err.KVs != nil && len(err.KVs) > 0
}

// String formatter for root errors.
func (err *ErrRoot) formatStr(format StringFormat) string {
	str := err.Msg + format.MsgStackSep
	if format.Options.WithTrace {
		stackArr := err.Stack.format(format.StackElemSep, format.Options.InvertTrace)
		for i, frame := range stackArr {
			str += format.PreStackSep + frame
			if i < len(stackArr)-1 {
				str += format.ErrorSep
			}
		}
	}
	return str
}

// JSON formatter for root errors.
func (err *ErrRoot) formatJSON(format JSONFormat) map[string]any {
	rootMap := make(map[string]any)
	rootMap["code"] = err.code.String()
	rootMap["message"] = err.Msg
	if err.HasKVs() {
		rootMap["KVs"] = err.KVs // TODO: debugging notes we lost the object at this point
	}
	if format.Options.WithTrace {
		rootMap["stack"] = err.Stack.format(format.StackElemSep, format.Options.InvertTrace)
	}
	return rootMap
}

// ErrLink represents a single error frame and the accompanying information.
type ErrLink struct {
	Msg   string
	Frame StackFrame
	code  Code
	KVs   map[string]any
}

// HasKVs returns true if the error has key-value pairs.
func (eLink *ErrLink) HasKVs() bool {
	return eLink.KVs != nil && len(eLink.KVs) > 0
}

// String formatter for wrap errors chains.
func (eLink *ErrLink) formatStr(format StringFormat) string {
	str := eLink.Msg + format.MsgStackSep
	if format.Options.WithTrace {
		str += format.PreStackSep + eLink.Frame.format(format.StackElemSep)
	}
	return str
}

// JSON formatter for wrap error chains.
func (eLink *ErrLink) formatJSON(format JSONFormat) map[string]any {
	wrapMap := make(map[string]any)
	wrapMap["code"] = eLink.code.String()
	wrapMap["message"] = fmt.Sprint(eLink.Msg)
	if eLink.HasKVs() {
		wrapMap["KVs"] = eLink.KVs
	}
	if format.Options.WithTrace {
		wrapMap["stack"] = eLink.Frame.format(format.StackElemSep)
	}
	return wrapMap
}
