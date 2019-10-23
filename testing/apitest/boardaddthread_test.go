package apitest

import (
	"api_test/threads"
	"testing"
)

func TestAddThreadCoreFields(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields()

	_ = test.Post("/thread").
		Form(op.Fields()).
		Expect(t).
		Status(201).
		Type("json").
		JSON(op.ExpectedResponse()).
		Done()

	op.Get().Thread("0").
		Check().IfEqualToExpectedThread()
}
