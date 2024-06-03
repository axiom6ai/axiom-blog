package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/user/model"
	"axiom-blog/internal/user/model/dao"
	"axiom-blog/internal/user/qo"
	"axiom-blog/internal/user/vo"
	jwt2 "axiom-blog/middleware/jwt"
	"axiom-blog/pkg/commonFunc/authCommonFunc"
	"axiom-blog/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"log"
	"regexp"
	"time"
)

type Users struct{}

type tokenInfo struct {
	ID    uuid.UUID
	name  string
	email string
	root  int
}

// 用户状态
const (
	one   int = iota + 1 //正常
	two                  //禁发文
	there                //冻结
)

// 生成token
func genToken(info tokenInfo) (token string, err error) {
	j := jwt2.NewJWT()

	// 构造用户claims信息(负荷)
	// 过期时间
	expiredTime := time.Now().Add(time.Duration(viper.GetInt("token.expires")) * time.Hour)
	claims := jwt2.CustomClaims{
		ID:    info.ID,
		Name:  info.name,
		Email: info.email,
		Root:  info.root,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredTime.Unix(),               // 过期时间
			IssuedAt:  time.Now().Unix(),                // 颁发时间
			Issuer:    viper.GetString("token.issuer"),  // 颁发者
			NotBefore: time.Now().Unix(),                // 生效时间
			Subject:   viper.GetString("token.subject"), // token主题
		},
	}
	token, err = j.CreateToken(claims)
	return token, err
}

// token添加进入redis
func addTokenToRedis(ctx *gin.Context, token string) (err error) {
	_, err = globalInit.RedisClient.SAdd(ctx, "token", token).Result()
	return err
}

// redis移除对应token
func removeTokenFromRedis(ctx *gin.Context, token string) (err error) {
	_, err = globalInit.RedisClient.SRem(ctx, "token", token).Result()
	return err
}

// 加密
func (u *Users) encryption(passwd string) (string, error) {
	store, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return string(store), err
	}
	return string(store), nil
}

// 校验密码
func (u *Users) comparePwd(storePasswd, passwd string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(storePasswd), []byte(passwd))
}

func (u *Users) Login(ctx *gin.Context) {
	loginQo := qo.LoginQO{}
	util.JsonConvert(ctx, &loginQo)
	users := new(dao.UserDAO).SelectByName(ctx, loginQo.Username)
	if 0 == len(users) {
		common.SendResponse(ctx, common.ErrUserNotFound, "")
		return
	}

	if u.comparePwd(users[0].Passwd, loginQo.Passwd) != nil {
		common.SendResponse(ctx, common.ErrPasswordIncorrect, "")
		return
	}
	//生成token
	tokenInfo := tokenInfo{
		users[0].ID,
		users[0].UserName,
		users[0].Email,
		users[0].IsRoot,
	}
	token, err := genToken(tokenInfo)
	if err != nil {
		common.SendResponse(ctx, common.ErrGenerateToken, err.Error())
		return
	}

	if ok := addTokenToRedis(ctx, token); ok != nil {
		common.SendResponse(ctx, common.ErrRedis, "")
		return
	}
	loginVo := vo.LoginVo{
		Token: token,
	}
	common.SendResponse(ctx, common.OK, loginVo)
}

func (u *Users) Logout(ctx *gin.Context) {
	token := ctx.Request.Header.Get("token")
	//清除redis对应token缓存
	err := removeTokenFromRedis(ctx, token)
	if err != nil {
		common.SendResponse(ctx, common.ErrRedis, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
}

func (u *Users) Register(ctx *gin.Context) {
	registerQO := qo.RegisterQO{}

	//校验必填请求参数
	util.JsonConvert(ctx, &registerQO)

	//密码限制8-32位字母、数字、特殊符号
	matched, _ := regexp.MatchString("^[\\Sa-zA-Z0-9]{8,32}", registerQO.Passwd)

	if !matched {
		log.Println("密码格式错误，匹配结果：", matched)
		common.SendResponse(ctx, common.ErrPasswordIncorrect, "")
		return
	}

	//校验唯一参数username、email
	users := new(dao.UserDAO).SelectByNameAndEmail(ctx, &model.User{UserName: registerQO.UserName, Email: registerQO.Email})
	if 0 < len(users) {
		common.SendResponse(ctx, common.ErrUserExisted, "")
		return
	} else {
		registerQO.State = global.ONE
		registerQO.IsRoot = global.ZERO
		storePasswd, err := u.encryption(registerQO.Passwd)
		if err != nil {
			common.SendResponse(ctx, common.ErrEncryption, err.Error())
			return
		}
		registerQO.Passwd = storePasswd
		user := model.User{}
		err = copier.Copy(&user, registerQO)

		hasRole, err := authCommonFunc.AuthCommonFunc{}.SelectRole(viper.GetString("defaultRole.ordinaryUser"))
		if !hasRole {
			if err == common.ErrRoleNotExisted {
				common.SendResponse(ctx, err, "用户添加角色失败！请联系管理员配置初始角色！")
				return
			}
			common.SendResponse(ctx, err, "")
			return
		}

		err = new(dao.UserDAO).Create(ctx, &user)
		if err != nil {
			common.SendResponse(ctx, common.ErrDatabase, err.Error())
			return
		}

		user1 := new(dao.UserDAO).SelectByName(ctx, registerQO.UserName)
		ID := user1[0].ID
		//生成token
		tokenInfo := tokenInfo{
			user1[0].ID,
			user1[0].UserName,
			user1[0].Email,
			user1[0].IsRoot,
		}
		token, err := genToken(tokenInfo)

		if err != nil {
			common.SendResponse(ctx, common.ErrGenerateToken, err.Error())
			return
		}

		//添加配置角色
		err = authCommonFunc.AuthCommonFunc{}.AddUserIntoRole(ID, viper.GetString("defaultRole.ordinaryUser"))
		if err != common.OK {
			common.SendResponse(ctx, err, "用户添加角色失败！请联系管理员配置初始角色！")
			return
		}

		loginVo := vo.LoginVo{
			Token: token,
			ID:    ID,
		}
		common.SendResponse(ctx, common.OK, loginVo)
		return
	}

}

func (u *Users) Info(ctx *gin.Context) {
	infoQO := qo.UserInfoQO{}
	util.JsonConvert(ctx, &infoQO)
	var user []model.User

	if infoQO.Email == "" && infoQO.Username == "" {
		common.SendResponse(ctx, common.ErrValidation, "")
		return
	} else if infoQO.Email == "" {
		user = new(dao.UserDAO).SelectByName(ctx, infoQO.Username)
	} else {
		user = new(dao.UserDAO).SelectByEmail(ctx, infoQO.Email)
	}
	common.SendResponse(ctx, common.OK, user[0])
}

func (u *Users) List(ctx *gin.Context) {
	//TODO 分页
	var users []model.User
	listQO := qo.UserListQO{}
	util.JsonConvert(ctx, &listQO)
	if listQO.State != global.ONE && listQO.State != global.TWO && listQO.State != global.THREE {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}
	globalInit.Db.Where("state", listQO.State).Find(&users)
	var userList []vo.UserListVO
	_ = copier.Copy(&userList, &users)
	common.SendResponse(ctx, common.OK, userList)
}

func (u *Users) Modify(ctx *gin.Context) {
	modifyQO := qo.ModifyQO{}
	util.JsonConvert(ctx, &modifyQO)
	user := model.User{}
	if err := copier.Copy(&user, modifyQO); err != nil {
		common.SendResponse(ctx, common.ErrBind, err.Error())
		return
	}

	//查询数据库中该username或email存在的用户信息
	userList := new(dao.UserDAO).SelectByNameAndEmail(ctx, &user)

	if len(userList) > 1 || (len(userList) == 1 && userList[0].ID != modifyQO.ID) {
		common.SendResponse(ctx, common.ErrUserExisted, "")
		return
	}

	encodePassWd, _ := u.encryption(user.Passwd)
	user.Passwd = encodePassWd

	if err := new(dao.UserDAO).UpdateUserInfo(ctx, &user); err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err.Error())
		return
	}

	common.SendResponse(ctx, common.OK, "修改成功")
	return
}
