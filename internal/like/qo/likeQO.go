package qo

// LikeQO 点赞/取消点赞参数不允许同时为空/存在，二选一
type LikeQO struct {
	//被点赞/取消点赞文章对象sn号
	Sn string `json:"sn"`

	//被点赞/取消点赞评论对象id
	CommentId int `json:"comment_id"`
}
