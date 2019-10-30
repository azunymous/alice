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
	e.POST("/thread").
		WithMultipart().WithFile("image", "image.png", op.WithImage()).WithForm(op.Fields()).
		Expect().
		Status(http.StatusCreated).JSON().Equal(op.Expected())

	op.Get().Thread(0).
		Check().ForThread(0).IfReply(1).NameIs("Anonymous")
}
