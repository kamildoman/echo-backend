package functions

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/kamildoman/echo-backend/storage"
	uuid "github.com/satori/go.uuid"
)

type Comment struct {
    ID string
    Message string `json:"message"`
    UserID string `json:"userId"`
	CreatedAt int `json:"created_at"`
    PostID string `json:"postId"`
    User User `gorm:"foreignkey:UserID"`
    Post Post `gorm:"foreignkey:PostID"`
}

type Post struct {
    ID string
    Message string `json:"message"`
	CreatedAt int `json:"created_at"`
    UserID string `json:"userId"`
    User User `gorm:"foreignkey:UserID"`
    Comments []Comment `gorm:"foreignkey:PostID"`
}

type Indexes struct {
	End int `json:"end"`
	Start int `json:"start"`
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

	//save post to database
	err = storage.DB.Create(&post).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create the post"})
		return err
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
		postMap := make(map[string]interface{})
		postMap["message"] = post.Message
		postMap["id"] = post.ID
		postMap["userId"] = post.UserID
		postMap["username"] = post.User.Username
		postMap["avatar"] = post.User.Avatar

		// Find all comments for the post
		var postComments []map[string]interface{}
		for _, comment := range post.Comments {
			commentMap := make(map[string]interface{})
			commentMap["message"] = comment.Message
			commentMap["userId"] = comment.UserID
			commentMap["username"] = comment.User.Username
			commentMap["avatar"] = comment.User.Avatar

			postComments = append(postComments, commentMap)
		}
		postMap["comments"] = postComments
		res[i] = postMap
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"data": res})

	return nil
}