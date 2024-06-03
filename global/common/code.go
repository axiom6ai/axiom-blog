package common

var (
	/*
		common errors
	*/
	OK                  = Errno{Code: 10000, Message: "OK"}
	InternalServerError = Errno{Code: 10001, Message: "Internal server error."}
	ErrBind             = Errno{Code: 10002, Message: "Error occurred while binding the request body to the struct."}
	ErrToken            = Errno{Code: 10003, Message: "Invalid Token."}
	ErrTokenExpired     = Errno{Code: 10004, Message: "Token is expired."}
	ErrTokenNotValidYet = Errno{Code: 10005, Message: "Token not active yet."}
	ErrTokenMalformed   = Errno{Code: 10006, Message: "That's not even a token."}
	ErrHandleToken      = Errno{Code: 10007, Message: "Couldn't handle this token:"}
	ErrGenerateToken    = Errno{Code: 10008, Message: "Generate Token is wrong."}
	ErrParam            = Errno{Code: 10009, Message: "Invalid Param."}
	ErrRedis            = Errno{Code: 10010, Message: "Redis Has Some Err."}

	/*
		system errors
	*/
	ErrValidation     = Errno{Code: 20001, Message: "Validation failed."}
	ErrDatabase       = Errno{Code: 20002, Message: "Database error."}
	ErrEncryption     = Errno{Code: 20003, Message: "encryption error."}
	ErrLoadPolicy     = Errno{Code: 20004, Message: "load policy error."}
	ErrorUploadFile   = Errno{Code: 20005, Message: "upload file error."}
	ErrorDownloadFile = Errno{Code: 20006, Message: "download file error."}

	/*
		person errors
	*/
	ErrUserNotFound           = Errno{Code: 20101, Message: "The user was not found."}
	ErrPasswordIncorrect      = Errno{Code: 20102, Message: "The password was incorrect."}
	ErrUserExisted            = Errno{Code: 20103, Message: "The user was existed."}
	ErrArticleNotExisted      = Errno{Code: 20104, Message: "The Article was not existed."}
	ErrRoleExisted            = Errno{Code: 20105, Message: "The role was existed."}
	ErrRoleNotExisted         = Errno{Code: 20106, Message: "The role was not existed."}
	ErrAddPermission          = Errno{Code: 20107, Message: "Add Permission failed."}
	ErrAccessDenied           = Errno{Code: 20108, Message: "Access denied."}
	ErrUserExistedInRole      = Errno{Code: 20109, Message: "The user already exists in the role."}
	ErrRelationshipNotExisted = Errno{Code: 20110, Message: "The relationship does not exist."}
	ErrRemovePermission       = Errno{Code: 20111, Message: "Remove Permission failed."}
	ErrUpdateUserAvatar       = Errno{Code: 20112, Message: "Update User Avatar failed."}
	ErrDeleteFile             = Errno{Code: 20113, Message: "Delete File failed."}
)
