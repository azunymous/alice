package threads

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/minio/minio-go/v6"
	"github.com/spf13/viper"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const boardID = "/obj/"

const (
	empty = -1

	undefined = 0

	adding = 1
	added  = 2

	prepare = 4
	getting = 5
	got     = 6

	assert = 7

	preparePost = 104
)

// TODO organize and split up adding, getting and asserting
type Controller struct {
	redis                   *redis.Client
	minio                   *minio.Client
	threadUnderModification Thread
	threadsFromDatabase     map[uint64]string
	threads                 map[uint64]Thread
	threadNoList            []uint64
	images                  map[uint64]*image.RGBA
	state                   int
	postingToThread         uint64
	postInDB                Post
	expectedReplyNumber     int
	formFields              map[string]string
	timeAtStart             time.Time
	// Pass through for latter WithFields() function for more readability. For actual expected thread No use passed in values.
	expectedPostNo uint64
	// Pass through for readability for selected thread
	lookingAtThreadNo uint64
}

func Operation() *Controller {
	return &Controller{
		redis:               redisClient(),
		minio:               minioClient(),
		threadsFromDatabase: map[uint64]string{},
		threads:             map[uint64]Thread{},
		images:              map[uint64]*image.RGBA{},
	}
}

// TODO Change to be parallelisable
func (tm *Controller) ClearRedis() *Controller {
	tm.redis.FlushAll()
	tm.state = empty
	return tm
}

func (tm *Controller) Add() *Controller {
	tm.state = adding
	return tm
}

func (tm *Controller) Get() *Controller {
	tm.state = getting
	return tm
}

func (tm *Controller) Thread(threadNo ...uint64) *Controller {
	switch tm.state {
	case adding:
		tm.threadUnderModification = thread()
		if len(threadNo) > 0 {
			tm.WithNo(threadNo[0])
		}
	case getting:
		get := tm.redis.Get(strconv.FormatUint(threadNo[0], 10))
		if get.Err() != nil {
			log.Fatalf("getting thread from redis error: %v", get.Err())
		}
		tm.threadsFromDatabase[threadNo[0]] = get.Val()
		tm.state = got
	}
	return tm
}

func (tm *Controller) And() *Controller {
	switch tm.state {
	case adding:
		tm.finalisedThread()
		tm.state = adding
	case getting:
		// TODO
		tm.state = getting
	}
	return tm
}

func (tm *Controller) AnotherThread() *Controller {
	tm.threadUnderModification = thread()
	tm.WithNo(1)
	return tm
}

func (tm *Controller) WithNo(no uint64) *Controller {
	tm.threadUnderModification.No = no
	tm.threadUnderModification.Timestamp = tm.threadUnderModification.Timestamp.Add(time.Duration(no * 1000000000))
	return tm
}

// TODO separate into more generalised function
func (tm *Controller) ToRedis() *Controller {
	if tm.state == adding {
		tm.finalisedThread()
		for _, t := range tm.threads {
			tm.redis.Set(boardID+":no", t.No+1, 0)
			tm.redis.ZAdd(boardID, redis.Z{Score: float64(t.Timestamp.UnixNano()), Member: t.No})
			tm.redis.Set(strconv.FormatUint(t.No, 10), t.AsJSON(), 0)
		}
	} else {
		panic("no thread was set up to be added to Redis")
	}
	return tm
}

// PrepareToPostThread initializes chain to create the required structs for posting a new thread including multipart form field and an image
func (tm *Controller) PrepareToPostThread(expectedThreadNo ...uint64) *Controller {
	tm.state = prepare
	tm.timeAtStart = time.Now()
	if len(expectedThreadNo) > 0 {
		tm.expectedPostNo = expectedThreadNo[0]
	}
	return tm
}

func (tm *Controller) PrepareToPostPost(expectedPostNo ...uint64) *Controller {
	tm.state = preparePost
	tm.timeAtStart = time.Now()
	if len(expectedPostNo) > 0 {
		tm.expectedPostNo = expectedPostNo[0]
	}
	return tm
}

func (tm *Controller) ToThread(threadNo uint64) *Controller {
	tm.postingToThread = threadNo
	return tm
}

func (tm *Controller) WithFields() *Controller {
	const comment = "Hello World!"
	values := map[string]string{"comment": comment}

	post := Post{
		No:      tm.expectedPostNo,
		Comment: comment,
	}

	if tm.state == prepare {
		tm.threads[tm.expectedPostNo] = Thread{
			Post: post,
		}
	} else {
		values["threadNo"] = strconv.FormatUint(tm.postingToThread, 10)
		threadToBeAddedTo := tm.threads[tm.postingToThread]
		threadToBeAddedTo.Replies = append(threadToBeAddedTo.Replies, post)
		tm.threads[tm.postingToThread] = threadToBeAddedTo
	}

	tm.formFields = values
	return tm
}

func (tm *Controller) WithImage() *bytes.Buffer {
	squareImage := image.NewRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{X: 100, Y: 100}})
	buffer := new(bytes.Buffer)
	_ = png.Encode(buffer, squareImage)
	tm.images[tm.expectedPostNo] = squareImage
	return buffer
}

func (tm *Controller) WithoutImage() *Controller {
	return tm
}

func (tm *Controller) WithNoName() *Controller {
	delete(tm.formFields, "name")
	return tm
}

func (tm *Controller) WithComment(comment string) *Controller {
	tm.formFields["comment"] = comment
	return tm
}

func (tm *Controller) Fields() map[string]string {
	return tm.formFields
}

func (tm *Controller) ExpectedResponse(threadNo ...uint64) string {
	switch tm.state {
	case added:
		response := BoardResponse{
			Status: "SUCCESS",
			No:     "0",
			Thread: tm.threads[0],
			Type:   "THREAD",
		}
		if len(threadNo) > 0 {
			response.No = strconv.FormatUint(tm.threads[threadNo[0]].No, 10)
			response.Thread = tm.threads[threadNo[0]]
		}
		return response.AsJSON()
	case prepare, preparePost:
		return `{"status":"SUCCESS","username":"","error":"","token":""}`
	}
	panic("nothing to expect")
}

func (tm *Controller) ExpectedArray() string {
	switch tm.state {
	case added:
		return "[" + strings.Join(tm.getAllThreadsAsJSON(), ",") + "]"
	}
	panic("nothing to expect")
}

func (tm *Controller) Expected(threadNo ...uint64) BoardResponse {
	switch tm.state {
	case added:
		response := BoardResponse{
			Status: "SUCCESS",
			No:     "0",
			Thread: tm.threads[0],
			Type:   "THREAD",
		}
		if len(threadNo) > 0 {
			response.No = strconv.FormatUint(tm.threads[threadNo[0]].No, 10)
			response.Thread = tm.threads[threadNo[0]]
		}
		return response
	case prepare, preparePost:
		return BoardResponse{Status: "SUCCESS"}
	}
	panic("nothing to expect")
}

func (tm *Controller) ExpectedThreads() []Thread {
	var list []Thread
	for i := len(tm.threadNoList) - 1; i >= 0; i-- {
		list = append(list, tm.threads[tm.threadNoList[i]])
	}
	return list
}

func (tm *Controller) getAllThreadsAsJSON() []string {
	var listAsJSON []string

	// Reverse order
	for i := len(tm.threadNoList) - 1; i >= 0; i-- {
		listAsJSON = append(listAsJSON, tm.threads[tm.threadNoList[i]].AsJSON())
	}
	return listAsJSON
}

func (tm *Controller) Check() *Controller {
	tm.state = assert
	return tm
}

func (tm *Controller) IfEqualToExpectedThread(threadNo ...uint64) *Controller {
	threadFromDBJSON := tm.threadsFromDatabase[threadNo[0]]
	threadFromDB := ThreadFromJSON(threadFromDBJSON)
	expectedThread := tm.threads[threadNo[0]]
	threadPostFromDB := threadFromDB.Post
	expectedThreadPost := expectedThread.Post
	tm.checkPostsAreAlmostEqual(threadPostFromDB, expectedThreadPost)
	return tm
}

func (tm *Controller) checkPostsAreAlmostEqual(post1 Post, expectedThreadPost Post) bool {
	if !tm.HasSameCoreFields(post1, expectedThreadPost) || !tm.hasTimeAdjusted(post1) || !tm.hasSameImage(post1, expectedThreadPost) {
		log.Fatalf("post not equal in core fields to expected. got %s expected %s", post1.AsJSON(), expectedThreadPost.AsJSON())
	}
	return true
}

func (tm *Controller) finalisedThread() {
	tm.threads[tm.threadUnderModification.No] = tm.threadUnderModification
	tm.threadNoList = append(tm.threadNoList, tm.threadUnderModification.No)
	defer func() { tm.threadUnderModification = Thread{} }()
	tm.state = added
}

// Check that the timestamp of the thread is after it was posted.
func (tm *Controller) hasTimeAdjusted(threadFromDB Post) bool {
	return threadFromDB.Timestamp.After(tm.timeAtStart)
}

func (tm *Controller) hasSameImage(t1 Post, t2 Post) bool {
	t1BucketAndFile := strings.Split(t1.Image, "/")
	t1Image, err1 := tm.minio.GetObject(t1BucketAndFile[0], t1BucketAndFile[1], minio.GetObjectOptions{})

	if err1 != nil {
		log.Fatalf("Error getting images for threads: thread 1 - %v", err1)
	}

	src, err := png.Decode(t1Image)
	if err != nil {
		log.Fatalf("Error decoding image %v", err)
	}
	b := src.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
	compare, _ := FastCompare(m, tm.images[t1.No])
	if compare == 0 {
		return true
	}

	return false
}

func FastCompare(img1, img2 *image.RGBA) (int64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := int64(0)

	for i := 0; i < len(img1.Pix); i++ {
		accumError += int64(sqDiffUInt8(img1.Pix[i], img2.Pix[i]))
	}

	return int64(math.Sqrt(float64(accumError))), nil
}

func sqDiffUInt8(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
}

func (tm *Controller) HasSameCoreFields(t1 Post, t2 Post) bool {
	if t1.No != t2.No || t1.Comment != t2.Comment {
		return false
	}
	return true
}

func (tm *Controller) IfNameIs(name string) *Controller {
	threadFromDB := ThreadFromJSON(tm.threadsFromDatabase[0])
	if threadFromDB.Name != name {
		log.Fatalf("Name of thread post got: %s wanted %s", threadFromDB.Name, name)
	}
	return tm
}

func (tm *Controller) IfCommentSegmentIs(segments []Segment) *Controller {
	threadFromDB := ThreadFromJSON(tm.threadsFromDatabase[0])
	if !reflect.DeepEqual(threadFromDB.CommentSegments, segments) {
		log.Fatalf("Comment segment for post got: %v, wanted: %v", threadFromDB.CommentSegments, segments)
	}
	return tm
}

func (tm *Controller) IfReply(no int) *Controller {
	thread := ThreadFromJSON(tm.threadsFromDatabase[tm.lookingAtThreadNo])
	if len(thread.Replies) < no {
		log.Fatalf("Thread %d does not have reply %d", tm.lookingAtThreadNo, no)
	}
	tm.postInDB = thread.Replies[no-1]
	tm.expectedReplyNumber = no - 1
	return tm
}

func (tm *Controller) ForThread(no uint64) *Controller {
	tm.lookingAtThreadNo = no
	return tm
}

func (tm *Controller) EqualToExpectedPost(no uint64) *Controller {
	tm.checkPostsAreAlmostEqual(tm.postInDB, tm.threads[tm.lookingAtThreadNo].Replies[tm.expectedReplyNumber])

	return tm
}

func redisClient() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: viper.GetString("redis.addr")})
}

func minioClient() *minio.Client {
	client, err := ConnectToMinio(viper.GetString("minio.addr"), viper.GetString("minio.access"), viper.GetString("minio.secret"))
	if err != nil {
		log.Fatalf("cannot connect to minio %v", err)
	}
	return client
}

func ConnectToMinio(addr, accessKey, secretAccessKey string) (*minio.Client, error) {
	minioClient, err := minio.New(addr, accessKey, secretAccessKey, false)
	return minioClient, err
}
