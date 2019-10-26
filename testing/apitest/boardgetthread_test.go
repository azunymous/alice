package apitest

import (
	"api_test/threads"
	"testing"
)

func TestGetThreadAllWithEmptyDB(t *testing.T) {
	threads.Operation().ClearRedis()

	_ = test.Get("/thread/all").
		Expect(t).
		Status(200).
		Type("json").
		JSON("null").
		Done()
}

func TestGetThreadAllWithOneThread(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().ToRedis()

	_ = test.Get("/thread/all").
		Expect(t).
		Status(200).
		Type("json").
		JSON(op.ExpectedArray()).
		Done()
}

func TestGetThreadAllWithTwoThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().And().AnotherThread().ToRedis()

	_ = test.Get("/thread/all").
		Expect(t).
		Status(200).
		Type("json").
		JSON(op.ExpectedArray()).
		Done()
}

func TestGetThreadAllWithTwoVariedNumberedThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(99).And().AnotherThread().WithNo(101).ToRedis()

	_ = test.Get("/thread/all").
		Expect(t).
		Status(200).
		Type("json").
		JSON(op.ExpectedArray()).
		Done()
}

func TestGetThreadWithThreadNotFound(t *testing.T) {
	threads.Operation().
		ClearRedis()

	_ = test.Get("/thread").AddQuery("no", "0").
		Expect(t).
		Status(404).
		Done()
}

func TestGetThreadWithThreadNotFoundWithOneThreadInDB(t *testing.T) {
	threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(0).ToRedis()

	_ = test.Get("/thread").AddQuery("no", "1").
		Expect(t).
		Status(404).
		Done()
}

func TestGetThreadWithOneThread(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(0).ToRedis()

	_ = test.Get("/thread").AddQuery("no", "0").
		Expect(t).
		Status(200).
		Type("json").
		JSON(op.ExpectedResponse()).
		Done()
}

func TestSecondThreadWithTwoThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().And().AnotherThread().WithNo(1).ToRedis()

	_ = test.Get("/thread").AddQuery("no", "1").
		Expect(t).
		Status(200).
		Type("json").
		JSON(op.ExpectedResponse(1)).
		Done()
}
