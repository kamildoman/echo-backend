package functions

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gofiber/fiber/v2"
)

func SendEmail(c *fiber.Ctx) error {
	conf := aws.Config{Region: aws.String("eu-central-1")}
	sess := session.New(&conf)

	svc := ses.New(sess)

	// Specify the email parameters
	sender := "admin@echoanalytics.pl" 
	recipient := "domanweb1@gmail.com"
	subject := "Test email"
	body := "This is a test email sent from Go using AWS SES."

	email := ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(recipient)},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data: aws.String(body),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Send the email using the SES service client
	_, err := svc.SendEmail(&email)
	if err != nil {
		panic(err)
	}

	c.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "avatar updated!"})
	return nil
}