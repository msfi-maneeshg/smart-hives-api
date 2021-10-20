package api

import (
	"fmt"
	"net/http"
	"smart-hives/api/common"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func isDeviceTypeExist(deviceType string) (status bool, err error) {
	url := common.IOT_URL + "device/types/" + deviceType
	resp, err := http.Get(url)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status, err
}

func isDeviceExist(deviceType, deviceID string) (status bool, err error) {
	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID
	resp, err := http.Get(url)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status, err
}

func isDestinationExist(deviceType string) (status bool, err error) {
	serviceID := "615a95d64a0b1217f089043c"
	url := common.IOT_URL + serviceID + "/destinations/" + deviceType
	resp, err := http.Get(url)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status, err
}

func createUserToken(objProfile FarmerProfileDetails) (objUserSession UserSession, err error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = objProfile.Username
	claims["email"] = objProfile.Email
	claims["exp"] = time.Now().Add(time.Minute * common.EXPIRE_TIME).Unix()

	objUserSession.UserToken, err = token.SignedString([]byte(common.MY_KEY))
	if err != nil {
		return objUserSession, err
	}

	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	objUserSession.RefereshToken, err = token.SignedString([]byte(common.REFERESH_KEY))
	if err != nil {
		return objUserSession, err
	}

	return objUserSession, nil
}

func CheckUserToken(r *http.Request) (objUserSession UserSession, err error) {
	var userToken string
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")

	if len(strArr) == 2 {
		userToken = strArr[1]
	}

	if userToken != "" {
		token, err := jwt.Parse(userToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%v", "There was an error")
			}
			return []byte(common.MY_KEY), nil
		})

		if err != nil {
			return objUserSession, err
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if token.Valid && ok {
			objUserSession.Username, _ = claims["username"].(string)
			objUserSession.Email, _ = claims["email"].(string)
			return objUserSession, nil
		}
		return objUserSession, fmt.Errorf("%v", "Invalid token")
	}

	return objUserSession, fmt.Errorf("%v", "Not Authorized")
}
