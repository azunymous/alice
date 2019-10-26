package apitest

import (
	"api_test/threads"
	"testing"
)

func TestAddThread(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields()

	_ = test.Post("/thread").
		Form(op.Fields()).
		Expect(t).
		Status(201).
		Type("json").
		JSON(op.ExpectedResponse()).
		Done()

	op.Get().Thread(0).
		Check().IfEqualToExpectedThread(0)
}

func TestAddThreadNumberIncreases(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread().WithNo(99).ToRedis().
		PrepareToPostThread(100).WithFields()

	_ = test.Post("/thread").
		Form(op.Fields()).
		Expect(t).
		Status(201).
		Type("json").
		JSON(op.ExpectedResponse()).
		Done()

	op.Get().Thread(100).
		Check().IfEqualToExpectedThread(100)
}

func TestAddThreadBlankNameIsAnonymous(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithNoName()

	_ = test.Post("/thread").
		Form(op.Fields()).
		Expect(t).
		Status(201).
		Type("json").
		JSON(op.ExpectedResponse()).
		Done()

	op.Get().Thread(0).
		Check().IfNameIs("Anonymous")
}

func TestAddThreadImageIsMandatory(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithoutImage()

	_ = test.Post("/thread").
		Form(op.Fields()).
		Expect(t).
		Status(400).
		Done()
}
