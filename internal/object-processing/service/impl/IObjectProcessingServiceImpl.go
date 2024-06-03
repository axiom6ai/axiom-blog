package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/internal/object-processing/model/dao"
	"axiom-blog/internal/object-processing/qo"
	"axiom-blog/internal/object-processing/vo"
	"axiom-blog/middleware/jwt"
	"axiom-blog/pkg/awsS3"
	"axiom-blog/pkg/commonFunc/userCommonFunc"
	"axiom-blog/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

/**
  @author: xichencx@163.com
  @date:2022/5/13
  @description:
**/

type ObjectProcessing struct{}

func tokenInfo(ctx *gin.Context) (Info *jwt.CustomClaims, err error) {
	Info, err = jwt.NewJWT().ParseToken(ctx.Request.Header.Get("token"))
	return
}

// 通过正则表达式判断文件后缀属于file、image、video、other
func fileTypeSuffix(ext string) string {
	regImg := regexp.MustCompile("^\\.(jpg|jpeg|png|gif|bmp|svg|webp)$")
	regVideo := regexp.MustCompile("^\\.(mp4|avi|flv|mkv|mov|mpg|mpeg|rm|rmvb|wmv|3gp|3g2)$")
	regFile := regexp.MustCompile("^\\.(doc|docx|xls|xlsx|ppt|pptx|pdf|txt|wps|et|dps|odt|ods|odp|pot|potx|pps|ppsx|sld|sldx|xlsb|xlsm|xlsx|xltm|xltx|xlam|xls)$")

	if regImg.MatchString(ext) {
		return "img"
	} else if regVideo.MatchString(ext) {
		return "video"
	} else if regFile.MatchString(ext) {
		return "file"
	} else {
		return "other"
	}
}

// 根据前缀判断并修改文件类型
func fileTypeByPrefix(file *multipart.FileHeader) {
	f, _ := file.Open()
	buf := make([]byte, file.Size)
	_, _ = f.Read(buf)
	ct := http.DetectContentType(buf)

	if ct == "application/octet-stream" {
		return
	}

	s := strings.Split(ct, "/")
	ext1 := "." + s[1]
	ext2 := path.Ext(file.Filename)

	if ext2 == "" {
		builder := strings.Builder{}
		builder.WriteString(file.Filename)
		builder.WriteString(ext1)
		file.Filename = builder.String()
		return
	}

	if ext1 != ext2 {
		strings.Replace(file.Filename, ext2, ext1, 1)
		return
	}

}

func getUserID(ctx *gin.Context) (userID uuid.UUID, err error) {
	token, err := tokenInfo(ctx)
	if err != nil {
		log.Println("获取tokenInfo失败：", err)
		return
	}
	userID = token.ID
	return
}

// 上传s3并生成数据库记录
func upload(ctx *gin.Context) (url string, err error) {
	file, err := ctx.FormFile("file")

	if file == nil {
		common.SendResponse(ctx, common.ErrParam, "file不能为空")
		return
	}

	//获取userID
	userID, err := getUserID(ctx)

	//根据文件前缀处理文件后缀
	fileTypeByPrefix(file)

	// 获取文件名
	filename := file.Filename

	//通过后缀获取文件类型
	ext := fileTypeSuffix(path.Ext(filename))

	//生成key(fileType/uid+timeString+filename)
	timeString := time.Now().Format("20060102150405")
	key := fmt.Sprintf("%s/%d%s%s", ext, userID, timeString, filename)
	storageKey := fmt.Sprintf("%d%s%s", userID, timeString, filename)

	f, _ := file.Open()
	buf := make([]byte, file.Size)
	_, err = f.Read(buf)

	//最小段5M
	var minPartSize int64
	minPartSize = 1024 * 1024 * 5

	//判断是否需要采用断点续传
	if file.Size <= minPartSize { //不需要断点续传
		err = awsS3.UploadFile(key, buf)
		if err != nil {
			log.Println("普通方式上传文件失败：", err)
			return "", err
		}
	} else { //断点续传
		err = awsS3.MultipartUpload(key, buf)
		if err != nil {
			log.Println("断点续传方式上传文件失败：", err)
			return "", err
		}
	}

	//获取URL
	url = awsS3.GetObjectUrl(key)

	//更新数据库
	fileDao := new(dao.File)

	//文件类型（1-图片，2-视频，3-文本，4-其他）
	var ft int

	if ext == "img" {
		ft = global.ONE
	} else if ext == "video" {
		ft = global.TWO
	} else if ext == "file" {
		ft = global.THREE
	} else {
		ft = global.FOUR
	}

	fileDao = &dao.File{
		Name:    storageKey,
		UserID:  userID,
		Address: url,
		Type:    ft,
		State:   global.ZERO,
	}

	_, err = fileDao.Creat()
	log.Println("数据库创建文件记录：", err)

	return url, err
}

func download(fileName string) (err error) {
	tmpPath := "/tmp/axiomroup/download"
	var file *os.File
	//检查是否存在临时文件夹，没有则创建临时文件夹
	if _, err = os.Stat(tmpPath); os.IsNotExist(err) {
		err = os.Mkdir(tmpPath, os.ModePerm)
	}

	//检查是否存在临时文件，没有则创建临时文件
	if _, err = os.Stat(tmpPath + fileName); os.IsNotExist(err) {
		file, e := os.Create(tmpPath + "/" + fileName)
		err = os.Chmod(file.Name(), os.ModePerm)

		if e != nil {
			err = e
			return
		}

		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				return
			}
		}(file)
	}

	err = awsS3.DownloadFile(fileName, file)
	if err != nil {
		//删除临时文件
		err = os.Remove(tmpPath + "/" + fileName)
	}
	return
}

func deleteFile(key string) (err error) {
	return awsS3.DeleteFile(key)
}

func (op ObjectProcessing) UploadAvatar(ctx *gin.Context) {
	//获取uid
	uid, err := getUserID(ctx)

	//上传头像并生成url
	url, err := upload(ctx)
	if url == "" {
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "上传用户头像失败!")
		return
	}

	if url != "" && err != nil {
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "新建用户用户记录失败!")
		return
	}

	//更新用户个人信息头像链接
	err = userCommonFunc.UserCommonFunc{}.UpdateUserAvatar(ctx, uid, url)

	if err != nil {
		log.Println("上传aws成功，数据库更新用户头像链接失败：", err.Error())
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "更新用户头像链接失败!")
		return
	}

	common.SendResponse(ctx, common.OK, vo.ObjectVO{Url: url})
	return
}

func (op ObjectProcessing) Upload(ctx *gin.Context) {

	//上传并生成url
	url, err := upload(ctx)
	if url == "" {
		log.Println("上传aws失败：", err.Error())
		common.SendResponse(ctx, common.ErrorUploadFile, "上传文件失败!")
		return
	}

	if url != "" && err != nil {
		log.Println("上传aws成功，生成上传记录失败：", err.Error())
		common.SendResponse(ctx, common.ErrorUploadFile, "新建上传记录失败!")
		return
	}

	common.SendResponse(ctx, common.OK, vo.ObjectVO{Url: url})
	return
}

func (op ObjectProcessing) Download(ctx *gin.Context, fileName string) {
	filePath := "/tmp/axiomroup/download/" + fileName
	err := download(fileName)

	if err != nil {
		e := common.ErrorDownloadFile
		e.Message += err.Error()
		common.SendResponse(ctx, e, "下载失败!")
		return
	}

	//返回文件
	fp, err := os.OpenFile(filePath, os.O_RDONLY, 4)
	if err != nil {
		log.Printf("文件：%s 打开失败", filePath)
	}
	defer func(fp *os.File) {
		_ = fp.Close()
	}(fp)

	bytes, err := ioutil.ReadAll(fp)

	if err != nil {
		log.Printf("读取文件：%s 失败", filePath)
	}

	common.SendResponse(ctx, common.OK, bytes)
	return
}

// 根据link获取key
func getKeyFromLink(link string) (key string) {
	s := strings.Split(link, "/")
	return s[len(s)-1]
}

func (op ObjectProcessing) UpdateAvatar(ctx *gin.Context) {
	query := new(qo.UpdateImgQO)
	util.JsonConvert(ctx, query)

	//根据头像link拿到key
	key := getKeyFromLink(query.ImgLink)

	//根据key删除头像
	err := deleteFile(key)

	if err != nil {
		log.Printf("s3删除%s记录失败，失败原因%v：", key, err)
	}

	//更新上传记录状态
	fileDao := new(dao.File)

	fileDao.Name = key
	fileDao.State = global.ONE
	err = fileDao.Update()
	if err != nil {
		log.Printf("更新头像%s记录失败，失败原因：%v", key, err)
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "")
		return
	}

	//上传头像,并创建数据库个人头像链接
	op.UploadAvatar(ctx)
}

func (op ObjectProcessing) Update(ctx *gin.Context) {
	query := new(qo.UpdateQO)
	util.JsonConvert(ctx, query)

	//根据link拿到key
	key := getKeyFromLink(query.Link)

	//根据key删除文件
	err := deleteFile(key)

	if err != nil {
		log.Printf("s3删除%s记录失败，失败原因%v：", key, err)
	}

	//更新上传记录状态
	fileDao := new(dao.File)

	fileDao.Name = key
	fileDao.State = global.ONE
	err = fileDao.Update()
	if err != nil {
		log.Printf("更新头像%s记录失败，失败原因：%v", key, err)
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "")
		return
	}

	//上传文件,创建数据库记录
	op.Upload(ctx)
}

func (op ObjectProcessing) DeleteFile(ctx *gin.Context) {
	query := new(qo.DeleteQO)
	util.JsonConvert(ctx, query)

	//根据link拿到key
	key := getKeyFromLink(query.Link)

	//根据key删除图片、文件等
	err := deleteFile(key)

	if err != nil {
		log.Println("删除失败：", err)
		common.SendResponse(ctx, common.ErrDeleteFile, "")
		return
	}

	//更新上传记录状态
	fileDao := new(dao.File)

	fileDao.Name = key
	fileDao.State = global.ONE
	err = fileDao.Update()
	if err != nil {
		log.Printf("更新头像%s记录失败，失败原因：%v", key, err)
		common.SendResponse(ctx, common.ErrUpdateUserAvatar, "")
		return
	}

	common.SendResponse(ctx, common.OK, "")
	return
}
