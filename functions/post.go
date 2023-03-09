package functions

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

type Comment struct {
    ID string
    Message string `json:"message"`
    UserID string `json:"userId"`
	CreatedAt time.Time `json:"created_at"`
    PostID string `json:"postId"`
    User User `gorm:"foreignkey:UserID"`
    Post Post `gorm:"foreignkey:PostID"`
}

type Post struct {
    ID string
    Message string `json:"message"`
	CreatedAt time.Time `json:"created_at"`
    UserID string `json:"userId"`
	Likes pq.StringArray `gorm:"type:text[]" json:"likes"`
    User User `gorm:"foreignkey:UserID"`
    Comments []Comment `gorm:"foreignkey:PostID"`
}

type Indexes struct {
	End int `json:"end"`
	Start int `json:"start"`
}

type Like struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}

type BroadcastMessage struct {
    Type string
    Data map[string]interface{}
	Ids []string
}


var broadcast = make(chan map[string]interface{})

func GetBroadcast() chan map[string]interface{} {
    return broadcast
}


func ToggleLike(context *fiber.Ctx) error {
    like := Like{}
	err := context.BodyParser(&like)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

    post := Post{}
    err = storage.DB.Where("id = ?", like.PostID).Preload("User").Preload("Comments").Preload("Comments.User").First(&post).Error
    if err != nil {
        context.Status(http.StatusBadRequest).JSON(
            &fiber.Map{"message": "could not find the post"})
        return err
    }

    foundIndex := -1
    for i, id := range post.Likes {
        if id == like.UserID {
            foundIndex = i
            break
        }
    }

    if foundIndex >= 0 {
        post.Likes = append(post.Likes[:foundIndex], post.Likes[foundIndex+1:]...)
    } else {
        post.Likes = append(post.Likes, like.UserID)
    }

    err = storage.DB.Save(&post).Error
    if err != nil {
        context.Status(http.StatusBadRequest).JSON(
            &fiber.Map{"message": "could not save the post"})
        return err
    }

	postData := serializePost(&post)
	message := BroadcastMessage{
		Type: "UPDATE_POST",
		Data: postData,
	}

	broadcast <- map[string]interface{}{
		"type": message.Type,
		"data": message.Data,
	}


    context.Status(http.StatusOK).JSON(
        &fiber.Map{"message": "post updated"})
    return nil
}

func CreatePost(context *fiber.Ctx) error {
	post := Post{}
	err := context.BodyParser(&post)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	//generate unique id
	id := uuid.NewV4().String()
	post.ID = id
	post.Likes = []string{}

	//save post to database
	err = storage.DB.Create(&post).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the post"})
		return err
	}

	postData := serializePost(&post)
	message := BroadcastMessage{
		Type: "NEW_POST",
		Data: postData,
	}

	broadcast <- map[string]interface{}{
		"type": message.Type,
		"data": message.Data,
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "post created"})
	return nil
}

func CreateComment(context *fiber.Ctx) error {
	comment := Comment{}
	err := context.BodyParser(&comment)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	//generate unique id
	id := uuid.NewV4().String()
	comment.ID = id

	//save comment to database
	err = storage.DB.Create(&comment).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the comment"})
		return err
	}

	post := Post{}
    err = storage.DB.Where("id = ?", comment.PostID).Preload("User").Preload("Comments").Preload("Comments.User").First(&post).Error

	postData := serializePost(&post)
	message := BroadcastMessage{
		Type: "UPDATE_POST",
		Data: postData,
	}

	broadcast <- map[string]interface{}{
		"type": message.Type,
		"data": message.Data,
	}
	

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "comment created"})
	return nil
}

func GetPosts(context *fiber.Ctx) error {
	indexes := Indexes{}
	err := context.QueryParser(&indexes)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	var posts []*Post

	db := storage.DB

	// Get all posts within the specified range and preload the related data
	db.Order("created_at desc").Limit(indexes.End - indexes.Start).Offset(indexes.Start).Preload("User").Preload("Comments").Preload("Comments.User").Find(&posts)

	// Construct the final response
	res := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		res[i] = serializePost(post)
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"data": res})

	return nil
}

func serializePost(post *Post) map[string]interface{} {
    postMap := make(map[string]interface{})
    postMap["message"] = post.Message
    postMap["id"] = post.ID
    postMap["likes"] = post.Likes
    postMap["userId"] = post.UserID
    postMap["created_at"] = post.CreatedAt
    postMap["username"] = post.User.Username
    postMap["avatar"] = post.User.Avatar

    // Find all comments for the post
    var postComments []map[string]interface{}
    for _, comment := range post.Comments {
        commentMap := make(map[string]interface{})
        commentMap["message"] = comment.Message
        commentMap["userId"] = comment.UserID
        commentMap["created_at"] = comment.CreatedAt
        commentMap["username"] = comment.User.Username
        commentMap["avatar"] = comment.User.Avatar

        postComments = append(postComments, commentMap)
    }
    postMap["comments"] = postComments

    return postMap
}
