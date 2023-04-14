package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"lab1/collections"
	"lab1/database"
	"net/http"
	"strings"
	"time"
)

type LoginForm struct {
	Username string `json:"user_name" form:"user_name"`
	Password string `json:"password" form:"password"`
	Remember bool   `json:"remember" form:"remember"`
}

var SECRET_KEY = []byte("aCSnbH6B1ATyRIDkOS3pB9xXMwOza9m7XrPnceNNVXxwvkbqjXwqgTuFgD1j6GsA")

func Register(c *gin.Context) {
	var (
		account             = collections.User{}
		db                  = database.GetMongoDB()
		upp, low, num, spec bool
		countChar           = 0
	)

	//Bind data
	if err := c.ShouldBindBodyWith(&account, binding.JSON); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Binding Error", nil)
		return
	}

	//Validate
	val := validate.Validate(
		&validators.StringIsPresent{Name: "Account", Field: account.UserName, Message: "Tên đăng nhập không được bỏ trống"},
		&validators.StringLengthInRange{Name: "Account", Field: account.UserName, Min: 5, Max: 254, Message: "Tên đăng nhập dài từ 4 đến 254 kí tự"},
		&validators.StringIsPresent{Name: "Password", Field: account.Password, Message: "Mật khẩu không được bỏ trống"},
		&validators.StringLengthInRange{Name: "Password", Field: account.Password, Min: 8, Message: "Mật khẩu phải dài từ 8 kí tự trở lên"},
		&validators.StringIsPresent{Name: "Password Confirm", Field: account.PasswordConfirm, Message: "Mật khẩu nhập lại không được bỏ trống"},
	)
	if val.HasAny() {
		ResponseError(c, http.StatusUnprocessableEntity, val.Errors[val.Keys()[0]][0], nil)
		return
	}

	for _, char := range account.Password {
		switch {
		case 'A' <= char && char <= 'Z':
			if upp == false {
				countChar++
			}
			upp = true
		case 'a' <= char && char <= 'z':
			if low == false {
				countChar++
			}
			low = true
		case '0' <= char && char <= '9':
			if num == false {
				countChar++
			}
			num = true
		default:
			if spec == false {
				countChar++
			}
			spec = true
		}
	}

	if countChar < 2 {
		ResponseError(c, http.StatusBadRequest, "Password bao gồm ít nhất 2 trong 4 loại: chữ hoa, chữ thường, số, ký tự đặc biệt", nil)
		return
	}

	//Confirm Password
	if account.Password != account.PasswordConfirm {
		ResponseError(c, http.StatusBadRequest, "Mật khẩu không khớp", nil)
		return
	}

	//Check exits account
	filter := bson.M{}

	if strings.Contains(account.UserName, "@") {
		filter["email"] = account.Email
	} else {
		filter["user_name"] = account.UserName
	}

	if existAccount := account.First(db, filter); existAccount == nil {
		ResponseError(c, http.StatusBadRequest, "Tài khoản đã tồn tại", nil)
		return
	}

	newPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, "Generate password error", nil)
		return
	}
	account.PasswordHash = string(newPassword)

	// Lưu dữ liệu
	if err = account.Create(db); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Đăng kí không thành công", nil)
		return
	}

	ResponseSuccess(c, http.StatusOK, "Đăng kí thành công!", nil)
	return
}

func Login(c *gin.Context) {
	var (
		request   LoginForm
		user      collections.User
		db        = database.GetMongoDB()
		start     = Now().UTC()
		expiredAt = start.Add(time.Hour * 1)
	)
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Binding error", nil)
		return
	}

	//Validate
	val := validate.Validate(
		&validators.StringIsPresent{Name: "Username", Field: request.Username, Message: "Tài khoản không được bỏ trống"},
		&validators.StringLengthInRange{Name: "Username", Field: request.Username, Min: 3, Max: 254, Message: "Tên tài khoản dài từ 3 đến 254 ký tự"},
		&validators.StringIsPresent{Name: "Password", Field: request.Password, Message: "Mật khẩu không được bỏ trống"},
	)
	if val.HasAny() {
		ResponseError(c, http.StatusUnprocessableEntity, val.Errors[val.Keys()[0]][0], nil)
		return
	}
	//Check exist account
	filterAccount := bson.M{
		"user_name":  request.Username,
		"deleted_at": nil,
	}
	if err := user.First(db, filterAccount); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Server error", err)
		return
	} else if err == mongo.ErrNoDocuments {
		ResponseError(c, http.StatusNotFound, "Tên đăng nhập không chính xác", nil)
		return
	}

	//Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		ResponseError(c, http.StatusUnauthorized, "Mật khẩu không chính xác", err)
		return
	}
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":  user.ID,
		"exp": expiredAt.Unix(),
	})
	if tokenSigned, err := token.SignedString(SECRET_KEY); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Server Error", nil)
		return
	} else {
		userToken := collections.UserToken{
			Token:      tokenSigned,
			ExpiredAt:  expiredAt,
			UserID:     user.ID,
			CreatedAt:  time.Now().UTC(),
			ModifiedAt: time.Now().UTC(),
		}
		if err = userToken.Create(db); err != nil {
			ResponseError(c, http.StatusInternalServerError, "Server Error", nil)
			return
		}
		ResponseSuccess(c, http.StatusOK, "Đăng nhập thành công!", gin.H{
			"token":      tokenSigned,
			"entry":      user,
			"expired_at": expiredAt,
		})
	}
}
