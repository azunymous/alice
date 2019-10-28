package apitest

import (
	"api_test/threads"
	"net/http"
	"testing"
)

func TestAddThread(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields()

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfEqualToExpectedThread(0)
}

func TestAddThreadNumberIncreases(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread().WithNo(99).ToRedis().
		PrepareToPostThread(100).WithFields()

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(100).
		Check().IfEqualToExpectedThread(100)
}

func TestAddThreadBlankNameIsAnonymous(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithNoName()

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfNameIs("Anonymous")
}

func TestAddThreadLinesAreParsedIntoSegmentsForEmptyPost(t *testing.T) {
	const comment = ``
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfCommentSegmentIs([]threads.Segment{{[]string{}, ""}})
}

func TestAddThreadLinesAreParsedIntoSegments(t *testing.T) {
	const comment = "Hello World\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfCommentSegmentIs([]threads.Segment{
		{[]string{}, "Hello World"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddThreadQuotesAreParsedIntoSegmentsWithQuoteFormat(t *testing.T) {
	const comment = ">This is a quote\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfCommentSegmentIs([]threads.Segment{
		{[]string{"quote"}, ">This is a quote"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddThreadNumberedQuotesAreParsedIntoSegmentsWithNumberedQuoteFormat(t *testing.T) {
	const comment = ">>0\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().IfCommentSegmentIs([]threads.Segment{
		{[]string{"noQuote"}, ">>0"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddThreadImageIsMandatory(t *testing.T) {
	op := threads.Operation().ClearRedis().
		PrepareToPostThread().WithFields().WithoutImage()

	e := setup(t)
	e.POST("/thread").
		WithMultipart().WithForm(op.Fields()).
		Expect().
		Status(http.StatusBadRequest)
}
