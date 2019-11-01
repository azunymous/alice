package apitest

import (
	"api_test/threads"
	"net/http"
	"testing"
)

func TestAddPost(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost(1).ToThread(0).WithFields()

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).EqualToExpectedPost(1)
}

func TestAddPostNumberIncreases(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost(1).ToThread(0).WithFields()

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.PrepareToPostPost(2).ToThread(0).WithFields()

	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(2).EqualToExpectedPost(2)
}

func TestAddPostWithTwoPosts(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread(0).WithPost(1).ToRedis().
		PrepareToPostPost(2).ToThread(0).WithFields()

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(2).EqualToExpectedPost(2)
}

func TestAddPostBlankNameIsAnonymous(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost(1).ToThread(0).WithFields().WithNoName()

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).NameIs("Anonymous")
}

func TestAddPostLinesAreParsedIntoSegmentsForEmptyPost(t *testing.T) {
	const comment = ``
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).IfCommentSegmentIs([]threads.Segment{{[]string{}, ""}})
}

func TestAddPostLinesAreParsedIntoSegments(t *testing.T) {
	const comment = "Hello World\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).IfCommentSegmentIs([]threads.Segment{
		{[]string{}, "Hello World"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddPostQuotesAreParsedIntoSegmentsWithQuoteFormat(t *testing.T) {
	const comment = ">This is a quote\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).IfCommentSegmentIs([]threads.Segment{
		{[]string{"quote"}, ">This is a quote"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddPostNumberedQuotesAreParsedIntoSegmentsWithNumberedQuoteFormat(t *testing.T) {
	const comment = ">>0\nNew Line\nAnother New Line"
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost().WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).IfCommentSegmentIs([]threads.Segment{
		{[]string{"noQuote"}, ">>0"},
		{[]string{}, "New Line"},
		{[]string{}, "Another New Line"},
	})
}

func TestAddPostNumberedQuotesAreAddedToPostsQuotedByList(t *testing.T) {
	const comment = ">>0\n>>1\nNew Line"
	op := threads.Operation().ClearRedis().
		Add().Thread(0).WithPost(1).ToRedis().
		PrepareToPostPost(2).WithFields().WithComment(comment)

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IsRepliedBy(2).
		Check().ForThread(0).IfReply(1).IsRepliedBy(2)
}

func TestAddPostImageIsOptional(t *testing.T) {
	op := threads.Operation().ClearRedis().
		Add().Thread(0).ToRedis().
		PrepareToPostPost().WithFields().WithoutImage()

	e := setup(t)
	e.POST("/post").
		WithMultipart().WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated)
}
