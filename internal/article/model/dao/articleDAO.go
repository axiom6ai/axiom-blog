package dao

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/article/model"
	"axiom-blog/internal/article/vo"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"strings"
)

type ArticleDAO struct {
	Aid     int   `json:"aid"`
	Sn      int64 `json:"sn"`
	Title   string
	UserID  uuid.UUID `json:"userID"`
	Cover   string
	Content string
	Tags    string
	State   int
	ViewNum bool `json:"view_num"`
	CmtNum  bool `json:"cmt_num"`
	ZanNum  bool `json:"zan_num"`
	Page    common.PageQO
}

var Db = &(globalInit.Db)

// 增加前后百分号
func addPercent(s string) string {
	builder := strings.Builder{}
	builder.WriteString("%")
	builder.WriteString(s)
	builder.WriteString("%")
	return builder.String()
}

func (ad ArticleDAO) CreatArticle(ctx *gin.Context, article *model.Article) (err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		if tx.Error != nil {
			return tx.Error
		}
		tx.Create(article)
		//文章扩展表插入一条记录
		tx.Error = ad.creatArticleEx(article.Sn)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
		return tx.Commit().Error
	}(tx)
	return err
}

func (ad ArticleDAO) SelectBySn(ctx *gin.Context, article *model.Article) *model.Article {
	(*Db).Model(&model.Article{}).Where("sn", article.Sn).First(&article)
	return article
}

func (ad ArticleDAO) SelectArticleBySn(sn int64) (a *model.Article) {
	(*Db).Where(model.Article{Sn: sn, State: global.ONE}).First(&a)
	return a
}

func (ad ArticleDAO) SelectArticleByState(state uint) (articlesMap map[int64]model.Article) {
	var articles []model.Article
	articles = []model.Article{}
	articlesMap = map[int64]model.Article{}

	globalInit.Db.Model(&model.Article{}).Where("state", state).Find(&articles)
	for _, v := range articles {
		articlesMap[v.Sn] = v
	}
	return articlesMap
}

func (ad ArticleDAO) FindArticles(ctx *gin.Context) (articlesVO vo.ArticleListVO) {
	tx := (*Db).WithContext(ctx).Model(&model.Article{})
	if ad.Page.PageNum > 0 && ad.Page.PageSize > 0 {
		size := ad.Page.PageSize
		num := ad.Page.PageNum
		tx = tx.Offset(size * (num - 1)).Limit(size).Order("aid asc")
	}
	if strings.Compare(ad.Page.Order, "desc") == 0 {
		tx = tx.Order("aid" + ad.Page.Order)
	}

	if ad.Sn != 0 { //sn精确搜索
		tx = tx.Where("axiom_blog_article.sn", ad.Sn)
	}
	if ad.Title != "" { //title模糊搜索
		tx = tx.Where("title Like ?", addPercent(ad.Title))
	}
	if ad.UserID != uuid.Nil { //uid精确搜索
		tx = tx.Where(`"user_ID" = ?`, ad.UserID)
	}
	if ad.Content != "" { //模糊搜索文章内容
		tx = tx.Where("content Like ？", addPercent(ad.Content))
	}
	if ad.Tags != "" {
		tx = tx.Where("tags In ?", strings.Split(ad.Tags, ","))
	}
	if ad.State >= 0 {
		tx = tx.Where("state", ad.State)
	}
	if ad.ViewNum {
		tx = tx.Order("view_num desc")
	}
	if ad.CmtNum {
		tx = tx.Order("cmt_num desc")
	}
	if ad.ZanNum {
		tx = tx.Order("zan_num desc")
	}
	tx, pageVO := ad.Page.NewPageVO(tx)
	articlesVO.PageVO = *pageVO
	row, err := tx.Select(`aid,axiom_blog_article.sn, title, user_ID, cover, content, tags, state, view_num, cmt_num, zan_num`).
		Joins("LEFT JOIN axiom_blog_article_ex ON axiom_blog_article.sn = axiom_blog_article_ex.sn ").Rows()

	defer func(row *sql.Rows) {
		err := row.Close()
		if err != nil {
			common.SendResponse(ctx, common.ErrDatabase, err)
		}
	}(row)

	log.Println(err)
	if err == nil {
		for row.Next() {
			article := &(vo.ArticleDetail{})
			err := tx.ScanRows(row, article)
			if err != nil {
				return vo.ArticleListVO{}
			}
			articlesVO.ArticleDetailList = append(articlesVO.ArticleDetailList, *article)
		}
	}
	return
}

func (ad ArticleDAO) creatArticleEx(sn int64) (err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		tx.Create(&model.ArticleEx{Sn: sn})
		log.Println("tx:", tx.Error)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
		return tx.Commit().Error
	}(tx)
	return
}

func (ad ArticleDAO) UpdateArticleEx(sn int64, view bool, cmt bool, zan bool, add bool) (err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		var ex model.ArticleEx
		tx.Where(&(model.ArticleEx{Sn: sn})).First(&ex)

		var updateFiled string
		var isAdd int
		if add {
			isAdd = global.ONE
		} else {
			isAdd = global.ZERO - global.ONE
		}

		if view {
			if !add && ex.ViewNum == global.ZERO {
				return nil
			}
			updateFiled = "view_num"
			ex.ViewNum += isAdd
		} else if cmt {
			if !add && ex.CmtNum == global.ZERO {
				return nil
			}
			updateFiled = "cmt_num"
			ex.CmtNum += isAdd
		} else if zan {
			if !add && ex.ZanNum == global.ZERO { //取消点赞时，防止出现负数
				return nil
			}
			updateFiled = "zan_num"
			ex.ZanNum += isAdd
		}

		tx.Select(updateFiled).Updates(&ex)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
		return tx.Commit().Error
	}(tx)
	return
}

func (ad ArticleDAO) UpdateArticle(ctx *gin.Context) (err error) {
	tx := globalInit.Transaction().Model(&model.Article{}).Where("sn", ad.Sn)
	err = func(db *gorm.DB) error {
		if tx.Error != nil {
			return tx.Error
		}
		tx.Omit("aid", "sn", `user_ID`).Update("state", ad.State).Updates(ad)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
		return tx.Commit().Error
	}(tx)
	return err
}
