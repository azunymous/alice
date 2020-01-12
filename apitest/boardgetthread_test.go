package apitest

import (
	"api_test/threads"
	"net/http"
	"testing"
)

func TestGetThreadAllWithEmptyDB(t *testing.T) {
	threads.Operation().ClearRedis()

	e := setup(t)
	e.GET("/thread/all").
		Expect().
		Status(http.StatusOK).JSON().Null()
}

func TestGetThreadAllWithOneThread(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().ToRedis()

	e := setup(t)
	e.GET("/thread/all").
		Expect().
		Status(http.StatusOK).JSON().Array().Equal(op.ExpectedThreads())
}

func TestGetThreadAllWithTwoThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().And().AnotherThread().ToRedis()

	e := setup(t)
	e.GET("/thread/all").
		Expect().
		Status(http.StatusOK).JSON().Array().Equal(op.ExpectedThreads())
}

func TestGetThreadAllWithTwoVariedNumberedThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(99).And().AnotherThread().WithNo(101).ToRedis()

	e := setup(t)
	e.GET("/thread/all").
		Expect().
		Status(http.StatusOK).JSON().Array().Equal(op.ExpectedThreads())
}

func TestGetThreadWithThreadNotFound(t *testing.T) {
	threads.Operation().
		ClearRedis()
	e := setup(t)
	e.GET("/thread").WithQuery("no", "0").
		Expect().
		Status(http.StatusNotFound)
}

func TestGetThreadWithThreadNotFoundWithOneThreadInDB(t *testing.T) {
	threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(0).ToRedis()

	e := setup(t)
	e.GET("/thread").WithQuery("no", "1").
		Expect().
		Status(http.StatusNotFound)
}

func TestGetThreadWithOneThread(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().WithNo(0).ToRedis()
	e := setup(t)
	e.GET("/thread").WithQuery("no", "0").
		Expect().
		Status(http.StatusOK).JSON().Equal(op.Expected())
}

func TestSecondThreadWithTwoThreads(t *testing.T) {
	op := threads.Operation().
		ClearRedis().
		Add().Thread().And().AnotherThread().WithNo(1).ToRedis()

	e := setup(t)
	e.GET("/thread").WithQuery("no", "1").
		Expect().
		Status(http.StatusOK).JSON().Equal(op.Expected(1))
}
