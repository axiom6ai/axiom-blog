package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/notify/model"
	"axiom-blog/internal/notify/model/dao"
	"axiom-blog/internal/notify/qo"
	"axiom-blog/internal/notify/vo"
	"axiom-blog/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"log"
	"sort"
	"strconv"
	"time"
)

/**
  @author: xichencx@163.com
  @date:2021/12/6
  @description: 通知服务接口实现
**/

type Notify struct{}

func (Notify) AddNotification(ctx *gin.Context, query *qo.AddNotificationQO) {
	//转换请求参数
	util.JsonConvert(ctx, query)

	//判断通知类型，若类型为4则userID默认为0;其余情况需将UID转换进行校验
	if query.Type == global.FOUR && len(query.UserID) > global.ONE {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	} else if query.Type == global.FOUR {
		query.UserID = []uuid.UUID{uuid.Nil}
	}

	//校验时间
	nowTime := time.Now()
	beginTimeInt, _ := strconv.ParseInt(query.BeginTime, 10, 64)
	endTimeInt, _ := strconv.ParseInt(query.EndTime, 10, 64)
	begin := time.Unix(beginTimeInt, 0)
	end := time.Unix(endTimeInt, 0)

	//校验参数
	if end.Before(begin) || end.Before(nowTime) {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	//校验参数时间段内是否存在通知
	notifyList, err := dao.Notify{}.GetNotifyList(begin, end)

	if err != nil {
		log.Println(err)
	}

	if len(notifyList) > global.ZERO {
		e := common.ErrParam
		e.Message = "该时间段内已存在通知!"
		common.SendResponse(ctx, e, "")
		return
	}

	//将uid列表、content、state、beginTime、endTime写入model
	notifyDao := new(dao.Notify)

	_ = copier.Copy(notifyDao, query)

	notifyDao.UserID = query.UserID
	notifyDao.Content = query.Content
	notifyDao.BeginTime = begin
	notifyDao.EndTime = end

	//model写入数据库
	id, err := notifyDao.Creat()

	if err != nil {
		common.SendResponse(ctx, err, "")
		return
	}
	common.SendResponse(ctx, common.OK, vo.AddNotificationVO{Id: id})
}

func (Notify) UpdateNotification(ctx *gin.Context, query *qo.UpdateNotificationQO) {
	//转换请求参数
	util.JsonConvert(ctx, query)

	//判断通知类型，若类型为4则uid默认为0;其余情况需将UID转换进行校验
	if query.Type == global.FOUR && len(query.UserID) > 1 {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	notifyDao := new(dao.Notify)

	//查询id是否存在记录
	result := globalInit.Db.Model(&model.Notify{}).Where("id", query.Id).Find(notifyDao)
	if result.RowsAffected == int64(global.ZERO) {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	//校验时间
	nowTime := time.Now()
	beginTimeInt, _ := strconv.ParseInt(query.BeginTime, 10, 64)
	endTimeInt, _ := strconv.ParseInt(query.EndTime, 10, 64)
	begin := time.Unix(beginTimeInt, 0)
	end := time.Unix(endTimeInt, 0)
	if end.Before(begin) || end.Before(nowTime) {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	//校验参数时间段内是否存在通知
	notifyList, err := dao.Notify{}.GetNotifyList(begin, end)

	if err != nil {
		log.Println(err)
	}

	if len(notifyList) > global.ZERO {
		e := common.ErrParam
		e.Message = "该时间段内已存在通知!"
		common.SendResponse(ctx, e, "")
		return
	}

	//将uid列表、content、state、beginTime、endTime写入model
	_ = copier.Copy(notifyDao, query)

	notifyDao.UserID = query.UserID
	notifyDao.Content = query.Content
	notifyDao.BeginTime = begin
	notifyDao.EndTime = end
	notifyDao.BeginTime = begin
	notifyDao.EndTime = end

	//更新
	err = notifyDao.Update(ctx)

	if err != nil {
		common.SendResponse(ctx, err, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
}

func (Notify) SystemNotify(ctx *gin.Context) {
	result := new(vo.SystemNotificationVO)
	//查询通知时间最近且为系统通知且处于开启状态的一条通知详情
	globalInit.Db.Model(&model.Notify{}).
		Where("type", global.FOUR).
		Where("state", global.ONE).
		Where("end_time > ?", time.Now()).
		Find(&result.NotificationList)
	if len(result.NotificationList) == global.ZERO {
		err := common.OK
		err.Message = "当前时间段暂无通知"
		common.SendResponse(ctx, err, "")
		return
	}

	//如果存在多条，则返回离当前时间最近的一条通知
	if len(result.NotificationList) > global.ONE {
		sort.Sort(result)
		log.Println("所有符合条件的通知：", result.NotificationList)
		common.SendResponse(ctx, common.OK, vo.SystemNotificationVO{NotificationList: result.NotificationList[:global.ONE]})
		return
	}

	common.SendResponse(ctx, common.OK, result)
	return
}
