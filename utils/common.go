package utils

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
)

type UserInfoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data User `json:"data"`
}

type UserListResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []User `json:"data"`
}

type User struct {
	UserID   int    `json:"user_id"`
	NickName string `json:"nick_name"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

func GetUserInfo(userList []int) map[int]interface{} {
	param, _ := json.Marshal(userList)
	url := beego.AppConfig.String("server_user") + "mid/user/nick_name"
	req := httplib.Get(url)
	req.Param("user_list", string(param))
	req.Header("Content-Type", "application/json;charset=utf-8")
	var user UserListResp
	_ = req.ToJSON(&user)
	// 组装用户信息
	var userInfoList = make(map[int]interface{})
	for _, u := range user.Data {
		var mobileAndEmail = make(map[string]string)
		mobileAndEmail["Email"] = u.Email
		mobileAndEmail["Mobile"] = u.Mobile
		userInfoList[u.UserID] = mobileAndEmail
	}
	logs.Info("userInfo: ", userInfoList)
	return userInfoList
}

func GetUserInfoAsyc(userList []int, userChan chan<- map[int]interface{}) {
	userChan <- GetUserInfo(userList)
	close(userChan)
}

func FindUserInfo(userInfo string) (user UserInfoResp) {
	url := beego.AppConfig.String("server_user") + "mid/user/information"
	req := httplib.Get(url)
	req.Param("accounts", userInfo)
	req.Header("Content-Type", "application/json;charset=utf-8")
	//var user UserInfoResp
	_ = req.ToJSON(&user)
	logs.Info("query user result: ", user)
	return
}

func FindUserInfoAsyc(userInfo string, userChan chan<- UserInfoResp) {
	userChan <- FindUserInfo(userInfo)
	close(userChan)
}

