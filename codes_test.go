package eris

import (
	"net/http"
	"testing"

	grpc "google.golang.org/grpc/codes"
)

func TestDefaultCodes(t *testing.T) {
	invalidHttp := Code(-1)
	resultCodeHttp := invalidHttp.ToHttp()
	if resultCodeHttp != http.StatusInternalServerError {
		t.Errorf("invalid code should be mapped to 500, but was %v", resultCodeHttp)
	}

	invalidGrpc := Code(-1)
	resultCodeGrpc := invalidGrpc.ToGrpc()
	if resultCodeGrpc != grpc.Unknown {
		t.Errorf("invalid code should be mapped to grpc.Unknown, but was %v", resultCodeGrpc)
	}
}

func TestInvalidConversions(t *testing.T) {
	code, validConversion := fromGrpc(grpc.OK)
	if validConversion {
		t.Errorf("grpc.Ok should not get converted to our error codes, but was converted to %v", code)
	}

	code, validConversion = fromHttp(200)
	if validConversion {
		t.Errorf("http 200 should not get converted to our error codes, but was converted to %v", code)
	}
}
