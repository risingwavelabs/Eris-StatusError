package eris_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/rotisserie/eris"
)

func TestUnpack(t *testing.T) {
	tests := map[string]struct {
		cause  error
		input  []string
		output eris.UnpackedError
	}{
		"nil error": {
			cause:  nil,
			input:  nil,
			output: eris.UnpackedError{},
		},
		"nil root error": {
			cause:  nil,
			input:  []string{"additional context"},
			output: eris.UnpackedError{},
		},
		"standard error wrapping with internal root cause (eris.New)": {
			cause: eris.New("root error").WithCode(eris.CodeUnknown),
			input: []string{"additional context", "even more context"},
			output: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "additional context",
					},
					{
						Msg: "even more context",
					},
				},
			},
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: eris.UnpackedError{
				ErrExternal: errors.New("external error"),
				ErrRoot: eris.ErrRoot{
					Msg: "additional context",
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "even more context",
					},
				},
			},
		},
		"no error wrapping with internal root cause (eris.Errorf)": {
			cause: eris.Errorf("%v", eris.CodeUnknown, "root error"),
			output: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
			},
		},
		"no error wrapping with external root cause (errors.New)": {
			cause: errors.New("external error"),
			output: eris.UnpackedError{
				ErrExternal: errors.New("external error"),
			},
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tt.cause, tt.input)
			if got := eris.Unpack(err); got.ErrChain != nil && tt.output.ErrChain != nil && !errChainsEqual(got.ErrChain, tt.output.ErrChain) {
				t.Errorf("Unpack() ErrorChain = %v, want %v", got.ErrChain, tt.output.ErrChain)
			}
			if got := eris.Unpack(err); !reflect.DeepEqual(got.ErrRoot.Msg, tt.output.ErrRoot.Msg) {
				t.Errorf("Unpack() ErrorRoot = %v, want %v", got.ErrRoot.Msg, tt.output.ErrRoot.Msg)
			}
		})
	}
}

func errChainsEqual(a []eris.ErrLink, b []eris.ErrLink) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Msg != b[i].Msg {
			return false
		}
	}

	return true
}

func TestFormatStr(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic root error": {
			input:  eris.New("root error").WithCode(eris.CodeUnknown),
			output: "code(unknown) root error",
		},
		"basic wrapped error": {
			input: eris.Wrap(
				eris.Wrap(
					eris.New("root error").WithCode(eris.CodeAlreadyExists),
					"additional context").WithCode(eris.CodeInvalidArgument),
				"even more context").WithCode(eris.CodeInvalidArgument),
			output: "code(invalid argument) even more context: code(invalid argument) additional context: code(already exists) root error",
		},
		"external wrapped error": {
			input:  eris.Wrap(errors.New("external error"), "additional context").WithCode(eris.CodeUnknown),
			output: "code(unknown) additional context: external error",
		},
		"external error": {
			input:  errors.New("external error"),
			output: "code(canceled) external error",
		},
		// This is the expected behavior, since this error does not hold any information
		"empty error": {
			input:  eris.New("").WithCode(eris.CodeUnknown),
			output: "",
		},
		"empty wrapped external error": {
			input:  eris.Wrap(errors.New(""), "additional context").WithCode(eris.CodeUnknown),
			output: "code(unknown) additional context: ",
		},
		"empty wrapped error": {
			input:  eris.Wrap(eris.New("").WithCode(eris.CodeUnknown), "additional context").WithCode(eris.CodeUnknown),
			output: "code(unknown) additional context: ",
		},
		// TODO: add tests for KVs and err code
	}
	for desc, tt := range tests {
		// without trace
		t.Run(desc, func(t *testing.T) {
			if got := eris.ToString(tt.input, false); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToString() got\n'%v'\nwant\n'%v'", got, tt.output)
			}
		})
	}
}

func TestInvertedFormatStr(t *testing.T) {

	tests := map[string]struct {
		input  error
		output string
	}{
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodeUnknown), "additional context").WithCode(eris.CodeUnknown), "even more context").WithCode(eris.CodeUnknown),
			output: "code(unknown) root error: code(unknown) additional context: code(unknown) even more context",
		},
		// TODO: Is this the expected behavior? Should an external error have a default code unknown?
		"external wrapped error": {
			input:  eris.Wrap(errors.New("external error"), "additional context").WithCode(eris.CodeUnknown),
			output: "external error: code(unknown) additional context",
		},
		"external error": {
			input:  errors.New("external error"),
			output: "external error code(canceled) ",
		},
		"empty wrapped external error": {
			input:  eris.Wrap(errors.New("some err"), "additional context").WithCode(eris.CodeUnknown),
			output: "some err: code(unknown) additional context",
		},
		"empty wrapped error": {
			input:  eris.Wrap(eris.New("err").WithCode(eris.CodeUnknown), "additional context").WithCode(eris.CodeUnknown),
			output: "code(unknown) err: code(unknown) additional context",
		},
	}
	for desc, tt := range tests {
		// without trace
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultStringFormat(eris.FormatOptions{
				InvertOutput: true,
				WithExternal: true,
			})
			if got := eris.ToCustomString(tt.input, format); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToString() got\n'%v'\nwant\n'%v'", got, tt.output)
			}
		})
	}
}

func TestFormatJSONwithKVs(t *testing.T) {
	nonSerializableObj := struct {
		a string
		b int
	}{
		a: "a",
		b: 1,
	}

	serializableObj := struct {
		A string `json:"a"`
		B int    `json:"b"`
	}{
		A: "aVal",
		B: 1,
	}

	// TODO: valid kvs with custom object that can be serialized
	// TODO: valid kvs with custom serializer for obj
	// TODO: Invalid kvs without serializer for obj

	// TODO: What about serializing pointers?

	tests := map[string]struct {
		input  error
		output string
	}{
		"basic root error + simple kvs": {
			input:  eris.New("root error").WithCode(eris.CodeCanceled).WithProperty("key", "value"),
			output: `{"root":{"KVs":{"key":"value"},"code":"canceled","message":"root error"}}`,
		},
		"basic wrapped error + kvs with objects": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodeNotFound).WithProperty("obj", nonSerializableObj), "additional context").WithCode(eris.CodeAlreadyExists).WithProperty("obj", serializableObj), "outer error"),
			output: `{"root":{"KVs":{"obj":{}},"code":"not found","message":"root error"},"wrap":[{"code":"internal","message":"outer error"},{"KVs":{"obj":{"a":"aVal","b":1}},"code":"already exists","message":"additional context"}]}`,
		},
		"basic wrapped error + kvs with objects 2": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodeNotFound).WithProperty("obj", serializableObj), "additional context"), "outer error"),
			output: `{"root":{"KVs":{"obj":{"a":"aVal","b":1}},"code":"not found","message":"root error"},"wrap":[{"code":"internal","message":"outer error"},{"code":"internal","message":"additional context"}]}`,
		},
		"basic wrapped error + ptr kvs": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithProperty("ptr", nil), "additional context"), "even more context"),
			output: `{"root":{"KVs":{"ptr":null},"code":"unknown","message":"root error"},"wrap":[{"code":"internal","message":"even more context"},{"code":"internal","message":"additional context"}]}`,
		},
		"external error + valid kvs": {
			input:  eris.Wrap(errors.New("external error"), "additional context").WithCode(eris.CodeNotFound),
			output: `{"external":"external error","root":{"code":"not found","message":"additional context"}}`,
		},
	}

	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			j := eris.ToJSON(tt.input, false)
			result, _ := json.Marshal(j)
			if got := string(result); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic root error": {
			input:  eris.New("root error").WithCode(eris.CodeCanceled),
			output: `{"root":{"code":"canceled","message":"root error"}}`,
		},
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodeNotFound), "additional context").WithCode(eris.CodeAlreadyExists), "even more context").WithCode(eris.CodeUnknown),
			output: `{"root":{"code":"not found","message":"root error"},"wrap":[{"code":"unknown","message":"even more context"},{"code":"already exists","message":"additional context"}]}`,
		},
		"external error": {
			input:  eris.Wrap(errors.New("external error"), "additional context").WithCode(eris.CodeDataLoss),
			output: `{"external":"external error","root":{"code":"data loss","message":"additional context"}}`,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			result, _ := json.Marshal(eris.ToJSON(tt.input, false))
			if got := string(result); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestInvertedFormatJSON(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodeAlreadyExists), "additional context").WithCode(eris.CodeUnknown), "even more context").WithCode(eris.CodeUnknown),
			output: `{"root":{"code":"already exists","message":"root error"},"wrap":[{"code":"unknown","message":"additional context"},{"code":"unknown","message":"even more context"}]}`,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultJSONFormat(eris.FormatOptions{
				InvertOutput: true,
			})
			result, _ := json.Marshal(eris.ToCustomJSON(tt.input, format))
			if got := string(result); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestFormatJSONWithStack(t *testing.T) {
	tests := map[string]struct {
		input      error
		rootOutput map[string]any
		wrapOutput []map[string]any
	}{
		"basic wrapped error": {
			input: eris.Wrap(eris.Wrap(eris.New("root error").WithCode(eris.CodePermissionDenied), "additional context").WithCode(eris.CodeUnavailable), "even more context").WithCode(eris.CodeUnknown),
			rootOutput: map[string]any{
				"code":    "permission denied",
				"message": "root error",
			},
			wrapOutput: []map[string]any{
				{"code": "unavailable", "message": "even more context"},
				{"code": "permission denied", "message": "additional context"},
			},
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultJSONFormat(eris.FormatOptions{
				WithTrace:   true,
				InvertTrace: true,
			})
			errJSON := eris.ToCustomJSON(tt.input, format)

			// make sure messages are correct and stack elements exist (actual stack validation is in stack_test.go)
			if rootMap, ok := errJSON["root"].(map[string]any); ok {
				if _, exists := rootMap["message"]; !exists {
					t.Fatalf("%v: expected a 'message' field in the output but didn't find one { %v }", desc, errJSON)
				}
				if rootMap["message"] != tt.rootOutput["message"] {
					t.Errorf("%v: expected { %v } got { %v }", desc, rootMap["message"], tt.rootOutput["message"])
				}
				if _, exists := rootMap["stack"]; !exists {
					t.Fatalf("%v: expected a 'stack' field in the output but didn't find one { %v }", desc, errJSON)
				}
			} else {
				t.Errorf("%v: expected root error is malformed { %v }", desc, errJSON)
			}

			// make sure messages are correct and stack elements exist (actual stack validation is in stack_test.go)
			if wrapMap, ok := errJSON["wrap"].([]map[string]any); ok {
				if len(tt.wrapOutput) != len(wrapMap) {
					t.Fatalf("%v: expected number of wrap layers { %v } doesn't match actual { %v }", desc, len(tt.wrapOutput), len(wrapMap))
				}
				for i := 0; i < len(wrapMap); i++ {
					if _, exists := wrapMap[i]["message"]; !exists {
						t.Fatalf("%v: expected a 'message' field in the output but didn't find one { %v }", desc, errJSON)
					}
					if wrapMap[i]["message"] != tt.wrapOutput[i]["message"] {
						t.Errorf("%v: expected { %v } got { %v }", desc, wrapMap[i]["message"], tt.wrapOutput[i]["message"])
					}
					if _, exists := wrapMap[i]["stack"]; !exists {
						t.Fatalf("%v: expected a 'stack' field in the output but didn't find one { %v }", desc, errJSON)
					}
				}
			} else {
				t.Errorf("%v: expected wrap error is malformed { %v }", desc, errJSON)
			}
		})
	}
}
