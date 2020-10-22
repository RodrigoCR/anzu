package feed

import (
	"github.com/tryanzu/core/modules/content"
	"github.com/tryanzu/core/modules/exceptions"

	//"github.com/tryanzu/core/modules/notifications"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
)

var lightPostFields bson.M = bson.M{"_id": 1, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "is_question": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}

type FeedModule struct {
	Errors       *exceptions.ExceptionsModule `inject:""`
	CacheService *goredis.Redis               `inject:""`
	User         *user.Module                 `inject:""`
	Content      *content.Module              `inject:""`
}

func (feed *FeedModule) Post(where interface{}) (post *Post, err error) {
	switch where := where.(type) {
	case bson.ObjectId, bson.M:
		var criteria = bson.M{"deleted_at": bson.M{"$exists": false}}

		switch where := where.(type) {
		case bson.ObjectId:
			criteria["_id"] = where
		case bson.M:
			for k, v := range where {
				criteria[k] = v
			}
		}

		// Use user feed reference to get the user and then create the user gaming instance
		err = deps.Container.Mgo().C("posts").Find(criteria).One(&post)
		if err != nil {
			err = exceptions.NotFound{Msg: "Invalid post id. Not found."}
			return
		}
	case *Post:
		post = where
	default:
		panic("Unkown argument")
	}

	post.SetDI(feed)
	return
}

func (feed *FeedModule) LightPost(post interface{}) (*LightPost, error) {

	switch post := post.(type) {
	case bson.ObjectId:

		scope := LightPostModel{}
		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").FindId(post).Select(lightPostFields).One(&scope)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid post id. Not found."}
		}

		post_object := &LightPost{data: scope, di: feed}

		return post_object, nil

	default:
		panic("Unkown argument")
	}
}

func (feed *FeedModule) LightPosts(posts interface{}) ([]LightPostModel, error) {

	switch posts := posts.(type) {
	case []bson.ObjectId:

		var list []LightPostModel

		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").Find(bson.M{"_id": bson.M{"$in": posts}}).Select(lightPostFields).All(&list)
		if err != nil {
			return nil, exceptions.NotFound{"Invalid posts id. Not found."}
		}

		return list, nil

	case bson.M:

		var list []LightPostModel

		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").Find(posts).Select(lightPostFields).All(&list)
		if err != nil {
			return nil, exceptions.NotFound{"Invalid posts criteria. Not found."}
		}

		return list, nil

	default:
		panic("Unkown argument")
	}
}

func (feed *FeedModule) GetComment(id bson.ObjectId) (comment *Comment, err error) {
	err = deps.Container.Mgo().C("comments").FindId(id).One(&comment)
	if err != nil {
		return
	}

	post, err := feed.Post(comment.PostId)
	if err != nil {
		return nil, err
	}

	comment.SetDI(post)
	return
}

func (feed *FeedModule) FulfillBestAnswer(list []LightPostModel) []LightPostModel {

	var ids []bson.ObjectId
	var comments []PostCommentModel

	for _, post := range list {

		// Generate the list of post id's
		ids = append(ids, post.Id)
	}

	database := deps.Container.Mgo()
	pipeline_line := []bson.M{
		{
			"$match": bson.M{"_id": bson.M{"$in": ids}, "solved": true},
		},
		{
			"$unwind": "$comments.set",
		},
		{
			"$match": bson.M{"comments.set.chosen": true},
		},
		{
			"$project": bson.M{"comment": "$comments.set"},
		},
	}

	pipeline := database.C("posts").Pipe(pipeline_line)
	err := pipeline.All(&comments)

	if err != nil {
		panic(err)
	}

	assoc := map[bson.ObjectId]PostCommentModel{}

	for _, comment := range comments {
		assoc[comment.Id] = comment
	}

	for index, post := range list {

		if comment, exists := assoc[post.Id]; exists {

			list[index].BestAnswer = &comment.Comment
		}
	}

	return list
}

func (feed *FeedModule) TrueCommentCount(id bson.ObjectId) int {
	var count int

	database := deps.Container.Mgo()
	count, err := database.C("comments").Find(bson.M{"post_id": id}).Count()

	if err != nil {
		panic(err)
	}

	return count
}
