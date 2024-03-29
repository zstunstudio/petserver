package controller

import (
	"net/http"
	"strconv"

	"github.com/chagspace/petserver/common"
	"github.com/chagspace/petserver/model"
	"github.com/chagspace/petserver/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "get_users",
	})
}

func GetUser(c *gin.Context) {
	user_id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.StatusBadRequestMessage("invalid user id"))
		return
	}

	user, isRecord := service.GetUserByUID(uint(user_id))
	if !isRecord {
		c.JSON(http.StatusOK, common.StatusOKMessage(gin.H{"user": nil}, "user not found"))
		return
	}
	c.JSON(http.StatusOK, common.StatusOKMessage(gin.H{"user": user}, "get user success"))
}

func CreateUser(c *gin.Context) {
	user := &model.UserModel{}
	c.BindJSON(&user)

	// check if username exists
	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   1,
			"msg":    "username is required",
			"status": "error",
		})
		return
	}

	// check if username is exists in database
	database_user, exist_user := service.GetUser(user.Username)
	if exist_user && database_user.Username == user.Username {
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"msg":    "username already exists",
			"status": "error",
		})
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(password)

	service.CreateUser(user)

	c.JSON(http.StatusOK, common.StatusOKMessage(gin.H{
		"uid":      user.UID,
		"username": user.Username,
	}, "create user success"))
}

func UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "update_user",
	})
}
func DeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "delete_user",
	})
}

// subscribe a user
func SubscribeUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "subscribe_user",
	})
}
func UnsubscribeUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "unsubscribe_user",
	})
}

// notify  a user
func NotifyUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"msg":     "success",
		"message": "notify_user",
	})
}

// Login user
func Login(c *gin.Context) {
	var user model.UserModel
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(
			http.StatusBadRequest,
			common.StatusBadRequestMessage("failed deserialization attempt, check request parameters"),
		)
		return
	}

	// check if username and password exists (should be validated in frontend and backend scheme)
	if user.Username == "" || user.Password == "" {
		c.JSON(
			http.StatusUnauthorized,
			common.StatusUnauthorizedMessage("username or password is incorrect"),
		)
		return
	}

	// check if user is already logged in and redirect to home page
	access_token, refresh_token, cookie_completed, _, _ := common.GetRenewableCookies(c)
	if cookie_completed {
		access_user_id, access_username, access_token_error := common.VerifyToken(access_token)
		_, _, refresh_token_error := common.VerifyToken(refresh_token)
		if access_token_error == nil && refresh_token_error == nil {
			// user is already logged in and has a valid token and refresh token
			c.JSON(http.StatusFound, common.StatusOKMessage(gin.H{
				"uid":      access_user_id,
				"username": access_username,
			}, "user is already logged in and has a valid token"))
			return
		}
	}

	// note:
	// because of the uniqueness of the user name, the password of the first user found is hashed with the current password
	database_user, allowed_user := service.GetUser(user.Username)
	if !allowed_user {
		c.JSON(
			http.StatusUnauthorized,
			common.StatusUnauthorizedMessage("username or password is incorrect"),
		)
		return
	}
	if common.VerifyPassword(database_user.Password, user.Password) != nil {
		c.JSON(http.StatusBadRequest, common.StatusBadRequestMessage("invalid account or password"))
		return
	}
	// set token to cookies
	isOk := common.UpdateStorageAuthToken(c, database_user.UID, user.Username)
	if !isOk {
		return
	}
	c.JSON(http.StatusOK, common.StatusOKMessage(gin.H{
		"uid":      database_user.UID,
		"username": user.Username,
	}, ""))
}

func Logout(c *gin.Context) {
	// jwt is strictly payload-dependent and belongs to the category of stateless services (the token cannot actually be destroyed after it is issued, it has to wait for time to expire)
	// The need to control logout is difficult to implement unless.
	// 1. use stateful storage (e.g. redis) to record the time of each token

	// delete access token and refresh token
	common.DeleteStorageAuthToken(c)

	// TODO: redirect to home page
	c.JSON(http.StatusOK, common.StatusOKMessage(gin.H{}, "logout success"))
}
