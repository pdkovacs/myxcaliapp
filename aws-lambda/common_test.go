package awslambda

import (
	"testing"
)

func TestCreateResponseWithChallange(t *testing.T) {
	respGot, errGot := createResponse(true, "", nil)
	if errGot != nil {
		t.Errorf("createResponseErr = %v; want nil", errGot)
	}
	headersGot := respGot.Headers
	challangeHeaderValue, gotChallangeHeader := headersGot["WWW-Authenticate"]
	if !gotChallangeHeader {
		t.Errorf("gotChallangeHeader = %v; want true", gotChallangeHeader)
	}
	if challangeHeaderValue != "Basic" {
		t.Errorf("challangeHeaderValue = %v; want \"Basic\"", challangeHeaderValue)
	}
	if respGot.StatusCode != 401 {
		t.Errorf("StatusCode = %v; want 401", respGot.StatusCode)
	}
}
