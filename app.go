package main

import (
	"fmt"
	"net/http"

	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	var cfg config
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	awsConfig := &aws.Config{Region: aws.String(cfg.AWS.Region)}

	r := gin.Default()

	r.POST("/register", func(c *gin.Context) {
		var request registerPostRequest
		if c.BindJSON(&request) == nil {
			awsSession, _ := session.NewSession()
			dbClient := dynamodb.New(awsSession, awsConfig)

			getInput := &dynamodb.GetItemInput{
				TableName: aws.String(cfg.AWS.UserTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(request.UserID),
					},
				},
				ConsistentRead: aws.Bool(true),
			}
			getResponse, err := dbClient.GetItem(getInput)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusInternalServerError, nil)
				return
			} else if len(getResponse.Item) > 0 {
				c.JSON(http.StatusBadRequest, nil)
				return
			}

			input := &dynamodb.UpdateItemInput{
				TableName: aws.String(cfg.AWS.UserTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(request.UserID),
					},
				},
				ExpressionAttributeNames: map[string]*string{
					"#Password": aws.String("Password"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":Password": {
						S: aws.String(request.Password),
					},
				},
				UpdateExpression: aws.String("set #Password = :Password"),
			}

			_, err = dbClient.UpdateItem(input)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusInternalServerError, nil)
				return
			}

			c.JSON(http.StatusOK, nil)
			return
		} else {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	})

	r.POST("/login", func(c *gin.Context) {
		var request loginPostRequest
		if c.BindJSON(&request) == nil {
			awsSession, _ := session.NewSession()
			dbClient := dynamodb.New(awsSession, awsConfig)

			getInput := &dynamodb.GetItemInput{
				TableName: aws.String(cfg.AWS.UserTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(request.UserID),
					},
				},
				AttributesToGet: []*string{
					aws.String("Password"),
				},
			}

			result, err := dbClient.GetItem(getInput)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusForbidden, nil)
				return
			}

			if request.Password != *result.Item["Password"].S {
				c.JSON(http.StatusForbidden, nil)
				return
			}

			sessionId, _ := uuid.NewUUID()

			updateInput := &dynamodb.UpdateItemInput{
				TableName: aws.String(cfg.AWS.UserTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(request.UserID),
					},
				},
				ExpressionAttributeNames: map[string]*string{
					"#SessionId": aws.String("SessionId"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":SessionId": {
						S: aws.String(sessionId.String()),
					},
				},
				UpdateExpression: aws.String("set #SessionId = :SessionId"),
			}

			_, err = dbClient.UpdateItem(updateInput)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusInternalServerError, nil)
				return
			}

			response := loginPostResponce{SessionID: sessionId.String()}

			c.JSON(http.StatusOK, response)
			return
		} else {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	})

	r.POST("/logout", func(c *gin.Context) {
		var request logoutPostRequest
		if c.BindJSON(&request) == nil {
			awsSession, _ := session.NewSession()
			dbClient := dynamodb.New(awsSession, awsConfig)

			if !isAuthorized(request.UserID, request.SessionID, dbClient, cfg) {
				c.JSON(http.StatusForbidden, nil)
				return
			}

			input := &dynamodb.UpdateItemInput{
				TableName: aws.String(cfg.AWS.UserTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(request.UserID),
					},
				},
				ExpressionAttributeNames: map[string]*string{
					"#SessionId": aws.String("SessionId"),
				},
				UpdateExpression: aws.String("remove #SessionId"),
			}

			_, err := dbClient.UpdateItem(input)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusInternalServerError, nil)
				return
			}

			c.JSON(http.StatusOK, nil)
		} else {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	})

	r.POST("/todo", func(c *gin.Context) {
		var request todoPostRequest

		if c.BindJSON(&request) == nil {
			awsSession, _ := session.NewSession()
			dbClient := dynamodb.New(awsSession, awsConfig)

			if !isAuthorized(request.UserID, request.SessionID, dbClient, cfg) {
				c.JSON(http.StatusForbidden, nil)
				return
			}

			todoID, _ := uuid.NewUUID()
			input := &dynamodb.UpdateItemInput{
				TableName: aws.String(cfg.AWS.TodoTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": {
						S: aws.String(todoID.String()),
					},
					"UserId": {
						S: aws.String(request.UserID),
					},
				},
				ExpressionAttributeNames: map[string]*string{
					"#Content": aws.String("Content"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":Content": {
						S: aws.String(request.Content),
					},
				},
				UpdateExpression: aws.String("set #Content = :Content"),
			}

			_, err := dbClient.UpdateItem(input)
			if err != nil {
				fmt.Println(err.Error())
				c.JSON(http.StatusInternalServerError, nil)
				return
			}

			response := todoPostResponse{TodoID: todoID.String()}

			c.JSON(http.StatusOK, response)
			return
		} else {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	})

	r.GET("/todo/:id", func(c *gin.Context) {
		todoID := c.Param("id")
		userID := c.Query("userid")
		sessionID := c.Query("sessionid")

		awsSession, _ := session.NewSession()
		dbClient := dynamodb.New(awsSession, awsConfig)

		if !isAuthorized(userID, sessionID, dbClient, cfg) {
			c.JSON(http.StatusForbidden, nil)
			return
		}

		input := &dynamodb.GetItemInput{
			TableName: aws.String(cfg.AWS.TodoTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String(todoID),
				},
				"UserId": {
					S: aws.String(userID),
				},
			},
			AttributesToGet: []*string{
				aws.String("Content"),
			},
		}

		result, err := dbClient.GetItem(input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		response := todoGetResponse{
			TodoID:  todoID,
			Content: *result.Item["Content"].S,
		}

		c.JSON(http.StatusOK, response)
		return
	})

	r.Run(":8090")
}

func isAuthorized(userID string, sessionID string, dbClient *dynamodb.DynamoDB, cfg config) bool {
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(cfg.AWS.UserTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(userID),
			},
		},
		AttributesToGet: []*string{
			aws.String("SessionId"),
		},
	}
	result, err := dbClient.GetItem(getInput)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return *result.Item["SessionId"].S == sessionID
}

type config struct {
	AWS awsConfig `json:"AWS"`
}

type awsConfig struct {
	Region        string `json:"Region"`
	UserTableName string `json:"UserTableName"`
	TodoTableName string `json:"TodoTableName"`
}

type authorizedRequest struct {
	UserID    string `json:"userid"`
	SessionID string `json:"sessionid"`
}

type registerPostRequest struct {
	UserID   string `json:"userid"`
	Password string `json:"password"`
}

type registerPostResponse struct {
	UserID string `json:"userid"`
}

type loginPostRequest struct {
	UserID   string `json:"userid"`
	Password string `json:"password"`
}

type loginPostResponce struct {
	SessionID string `json:"sessionid"`
}

type logoutPostRequest struct {
	authorizedRequest
}

type todoPostRequest struct {
	authorizedRequest
	Content string `json:"content"`
}

type todoPostResponse struct {
	TodoID string `json:"todoid"`
}

type todoGetResponse struct {
	TodoID  string `json:"todoid"`
	Content string `json:"content"`
}

type todo struct {
	TodoID  string `json:"todoid"`
	Content string `json:"content"`
}
