package qo

/**
  @author: ethan.chen@axiomroup.cn
  @date:2021/10/11
  @description:删除评论请求参数
**/

type DeleteCommentQO struct {
	//评论ID
	CommentId int `binding:"required"`
}

type DeleteCommentReplyQO struct {
	//评论回复ID
	Id int `binding:"required"`
}
