package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/internal/auth"
	"axiom-blog/internal/auth/qo"
	"axiom-blog/internal/auth/vo"
	"axiom-blog/pkg/commonFunc/userCommonFunc"
	"axiom-blog/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"strings"
)

type Auth struct{}

// 校验角色是否存在
func checkRoles(ctx *gin.Context, roles []string) (roleMap map[string]bool) {
	roleMap = make(map[string]bool)
	e, _ := auth.GetE(ctx)
	for _, v := range roles {
		hasGroup := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+v)
		if len(hasGroup) == global.ZERO {
			roleMap[v] = false
			continue
		}
		roleMap[v] = true
	}
	return
}

// 校验用户是否存在
func checkUsers(ctx *gin.Context, users []uuid.UUID) (userMap map[uuid.UUID]bool, err error) {
	userMap = make(map[uuid.UUID]bool)
	userModelMap := userCommonFunc.IUser(userCommonFunc.UserCommonFunc{}).FindUser(ctx, users, "", "")
	for _, v := range users {
		if userInfo, ok := userModelMap[v]; !ok || int(userInfo.State) != global.ONE {
			userMap[v] = false
			continue
		}
		userMap[v] = true
	}
	return
}

// AllPolicies 查询所有权限
func (a Auth) AllPolicies(ctx *gin.Context) {
	e, _ := auth.GetE(ctx)
	permission := e.GetNamedGroupingPolicy("g2")
	log.Println(permission)
	permissionMap := map[string]string{}
	for _, v := range permission {
		_, ok := permissionMap[v[1]]
		if !ok {
			permissionMap[v[1]] = v[0]
		}
	}
	common.SendResponse(ctx, common.OK, permissionMap)
	return
}

// AllRoles 查询所有角色及其权限
func (a Auth) AllRoles(ctx *gin.Context) {
	e, _ := auth.GetE(ctx)

	//角色
	role := e.GetNamedGroupingPolicy("g")
	roleMap := map[string][]string{}
	for _, v := range role {
		_, ok := roleMap[strings.TrimPrefix(v[1], global.RolePrefix)]
		if !ok {
			roleMap[strings.TrimPrefix(v[1], global.RolePrefix)] = []string{}
		}
	}
	log.Println("角色map:", roleMap)

	//权限
	permission := e.GetNamedGroupingPolicy("g2")
	permissionMap := map[string]string{}
	for _, v := range permission {
		_, ok := permissionMap[v[1]]
		if !ok {
			permissionMap[v[1]] = v[0]
		}
	}
	log.Println("权限map:", permissionMap)

	//角色-权限表关系
	roleAndPermission := e.GetPolicy()
	log.Println(roleAndPermission)
	for _, v := range roleAndPermission {
		rName := strings.TrimPrefix(v[0], global.RolePrefix)
		roleMap[rName] = append(roleMap[rName], permissionMap[v[1]])
	}
	log.Println("角色与权限关系map:", roleMap)

	common.SendResponse(ctx, common.OK, roleMap)
	return
}

// AddPermission 系统添加单个权限
func (a Auth) AddPermission(ctx *gin.Context) {
	p := qo.PermissionQO{}
	e, _ := auth.GetE(ctx)
	util.JsonConvert(ctx, &p)

	ok, err := e.AddNamedGroupingPolicy("g2", p.Uri, p.PName)
	if err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err)
		return
	}
	//校验数据库是否存在该条权限规则
	if !ok {
		common.SendResponse(ctx, common.OK, "数据库存在该接口权限规则:"+p.Uri)
		return
	}
	common.SendResponse(ctx, common.OK, "接口权限添加成功！")
	return
}

// AddRole 添加角色
func (a Auth) AddRole(ctx *gin.Context) {
	name := new(qo.AddRoleQO)
	util.JsonConvert(ctx, name)
	e, _ := auth.GetE(ctx)
	result, err := e.AddNamedGroupingPolicy("g", "", global.RolePrefix+name.RName)
	if err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err)
		return
	}
	if !result {
		common.SendResponse(ctx, common.ErrRoleExisted, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// AddPermissionsForRole 角色添加权限
func (a Auth) AddPermissionsForRole(ctx *gin.Context) {
	e, _ := auth.GetE(ctx)
	gap := new(qo.GroupAddPermissionQO)
	util.JsonConvert(ctx, gap)

	if len(gap.PName) < 1 || gap.RName == "" {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	//policyList := strings.Join(gap.PName,", ")
	//log.Println(policyList)
	//校验角色
	hasGroup := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+gap.RName)
	if len(hasGroup) == 0 {
		common.SendResponse(ctx, common.ErrRoleNotExisted, "")
		return
	}
	//根据PName查询策略字段
	//hasPermission := e.GetFilteredNamedGroupingPolicy("g2", 1, policyList)
	//log.Println(hasPermission)

	var failureString string
	for _, v := range gap.PName {
		hasPolicy, _ := e.AddPolicy(global.RolePrefix+gap.RName, v, global.Operate)
		if !hasPolicy {
			var build strings.Builder
			build.WriteString(failureString)
			build.WriteString(v)
			build.WriteString(" ")
			failureString = build.String()
		}
	}
	if failureString != "" {
		common.SendResponse(ctx, common.ErrAddPermission, "添加失败的权限为："+failureString)
		return
	}
	common.SendResponse(ctx, common.OK, "权限添加成功")
	return
	//if len(hasPermission) > 0 {
	//	hasPolicy, err := e.AddPolicy(axiomConst.RolePrefix+gap.RName, gap.PName, axiomConst.Operate)
	//	if err != nil {
	//		common.SendResponse(ctx, common.ErrDatabase, "添加失败"+err.Error())
	//		return
	//	}
	//	if !hasPolicy {
	//		common.SendResponse(ctx, common.OK, "该角色已存在该权限")
	//		return
	//	}
	//	common.SendResponse(ctx, nil, "添加成功")
	//	return
	//} else {
	//	common.SendResponse(ctx, common.ErrAddPermission, "")
	//	return
	//}
}

// RemovePermissionsFromRole 角色移除权限
func (a Auth) RemovePermissionsFromRole(ctx *gin.Context) {
	query := new(qo.DeletePermissionFromRoleQO)
	util.JsonConvert(ctx, query)
	e, _ := auth.GetE(ctx)
	//校验角色是否存在
	role := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+query.RName)
	if len(role) < global.ONE {
		common.SendResponse(ctx, common.ErrRoleNotExisted, "")
		return
	}
	//校验权限是否为空、重复
	if len(query.PName) < global.ONE {
		common.SendResponse(ctx, common.ErrParam, "权限不能为空！")
		return
	}
	var permissionMap map[string]bool
	for _, v := range query.PName {
		if _, ok := permissionMap[v]; ok {
			common.SendResponse(ctx, common.ErrParam, "权限参数存在重复值！")
			return
		}
	}

	//解除权限-角色关联
	for _, v := range query.PName {
		_, _ = e.RemoveFilteredNamedPolicy("p", 0, global.RolePrefix+query.RName, v)
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// AddUserIntoRole 添加用户-角色关联
func (a Auth) AddUserIntoRole(ctx *gin.Context) {
	userIntoGroup := new(qo.AddUserIntoRoleQO)
	util.JsonConvert(ctx, userIntoGroup)
	e, _ := auth.GetE(ctx)

	//TODO 校验uid

	//校验角色是否存在
	hasGroup := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+userIntoGroup.RName)
	if len(hasGroup) == 0 {
		common.SendResponse(ctx, common.ErrRoleNotExisted, "")
		return
	}

	result, err := e.AddRoleForUser(global.UserPrefix+userIntoGroup.UserID.String(), global.RolePrefix+userIntoGroup.RName)
	if err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err)
		return
	}
	if !result {
		common.SendResponse(ctx, common.ErrUserExistedInRole, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// DeletePermission 移除权限，且解除权限-角色关联
func (a Auth) DeletePermission(ctx *gin.Context) {
	query := new(qo.DeletePermissionQO)
	util.JsonConvert(ctx, query)
	e, _ := auth.GetE(ctx)
	result, err := e.RemoveFilteredNamedGroupingPolicy("g2", 1, query.PName)
	if err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err)
		return
	}
	if !result {
		common.SendResponse(ctx, common.ErrRemovePermission, "")
		return
	}
	_, err = e.RemoveFilteredNamedPolicy("p", 1, query.PName)
	common.SendResponse(ctx, common.OK, err)
	return
}

// DeleteRole 删除角色，且解除角色与权限关联及角色与用户关联
func (a Auth) DeleteRole(ctx *gin.Context) {
	query := new(qo.DeleteRoleQO)
	util.JsonConvert(ctx, query)
	e, _ := auth.GetE(ctx)
	//result, err := e.RemoveFilteredNamedGroupingPolicy("g", 1, axiomConst.RolePrefix+query.RName)
	//if !result {
	//	common.SendResponse(ctx, common.ErrRoleNotExisted, "")
	//	return
	//}
	//if err != nil {
	//	common.SendResponse(ctx, common.ErrDatabase, err)
	//	return
	//}
	//_, _ = e.RemoveFilteredNamedPolicy("p", 0, axiomConst.RolePrefix+query.RName)
	//common.SendResponse(ctx, common.OK, "")

	if len(query.RName) < global.ONE {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}
	for _, v := range query.RName {
		_, _ = e.RemoveFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+v)
		_, _ = e.RemoveFilteredNamedPolicy("p", 0, global.RolePrefix+v)
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// GetUserRoles 查询用户角色
func (a Auth) GetUserRoles(ctx *gin.Context) {
	query := new(qo.GetUserRolesQO)
	util.JsonConvert(ctx, query)
	e, _ := auth.GetE(ctx)
	var userRolesVo []vo.UserRolesVO

	if len(query.UserID) < global.ZERO {
		common.SendResponse(ctx, common.ErrParam, "")
		return
	}

	uMap := make(map[uuid.UUID]bool)
	for _, v := range query.UserID {
		if _, ok := uMap[v]; ok {
			common.SendResponse(ctx, common.ErrParam, "userID存在重复!")
			return
		}
		uMap[v] = true
	}

	getRole := func(s [][]string) (res []string) {
		for _, v := range s {
			res = append(res, strings.TrimPrefix(v[global.ONE], global.RolePrefix))
		}
		return res
	}

	for _, v := range query.UserID {
		userRoles := e.GetFilteredNamedGroupingPolicy("g", global.ZERO, global.UserPrefix+v.String())
		log.Println(userRoles)
		userRolesInfo := vo.UserRolesVO{
			UserID:    v,
			RoleNames: getRole(userRoles),
		}
		userRolesVo = append(userRolesVo, userRolesInfo)
	}
	userMap := userCommonFunc.UserCommonFunc{}.FindUser(ctx, query.UserID, "", "")

	for k, v := range userRolesVo {
		userRolesVo[k].UserName = userMap[v.UserID].Nickname
	}

	common.SendResponse(ctx, common.OK, userRolesVo)
	return
}

// RoleRemoveUser 用户移除角色
func (a Auth) RoleRemoveUser(ctx *gin.Context) {
	query := new(qo.DeleteUserRoleQO)
	util.JsonConvert(ctx, query)
	e, _ := auth.GetE(ctx)
	result, err := e.RemoveFilteredNamedGroupingPolicy("g", 0,
		global.UserPrefix+query.UserID.String(), global.RolePrefix+query.RName)

	if !result {
		common.SendResponse(ctx, common.ErrRelationshipNotExisted, "")
		return
	}
	if err != nil {
		common.SendResponse(ctx, common.ErrDatabase, err)
		return
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// UserAddRolesInBatches 用户批量添加角色
func (a Auth) UserAddRolesInBatches(ctx *gin.Context) {
	query := new(qo.UserAddRolesInBatches)
	util.JsonConvert(ctx, query)

	//校验参数
	if global.ONE > len(query.RName) {
		common.SendResponse(ctx, common.ErrParam, "角色不能为空！")
		return
	}

	// 校验用户
	userMap, _ := checkUsers(ctx, []uuid.UUID{query.UserID})
	if !userMap[query.UserID] {
		common.SendResponse(ctx, common.ErrParam, "用户不存在！")
		return
	}

	//校验角色
	roleListVo := vo.Roles{}
	roleMap := checkRoles(ctx, query.RName)
	log.Println("不存在的角色map：", roleMap)
	for k, v := range roleMap {
		if !v {
			roleListVo.RoleName = append(roleListVo.RoleName, k)
		}
	}
	log.Println("不存在的角色列表：", roleListVo.RoleName)
	if global.ZERO < len(roleListVo.RoleName) {
		e := common.ErrParam
		e.Message = "角色不存在！"
		common.SendResponse(ctx, e, roleListVo)
		return
	}

	e, _ := auth.GetE(ctx)
	for _, v := range query.RName {
		result, err := e.AddRoleForUser(global.UserPrefix+query.UserID.String(), global.RolePrefix+v)
		if !result {
			log.Printf("用户\"%s\"添加角色：%s失败,失败原因：\"%v\" ", query.UserID.String(), v, err)
		}
	}
	common.SendResponse(ctx, common.OK, "")
	return
}

// UserDeleteRolesInBatches 批量删除用户的角色
func (a Auth) UserDeleteRolesInBatches(ctx *gin.Context) {
	query := new(qo.UserDeleteRolesInBatches)
	util.JsonConvert(ctx, query)

	//校验参数
	if global.ONE > len(query.RName) {
		common.SendResponse(ctx, common.ErrParam, "角色不能为空！")
		return
	}

	// 校验用户
	userMap, _ := checkUsers(ctx, []uuid.UUID{query.UserID})
	if !userMap[query.UserID] {
		common.SendResponse(ctx, common.ErrParam, "用户不存在！")
		return
	}

	//校验角色
	roleListVo := vo.Roles{}
	roleMap := checkRoles(ctx, query.RName)
	for k, v := range roleMap {
		if !v {
			roleListVo.RoleName = append(roleListVo.RoleName, k)
		}
	}
	if global.ZERO < len(roleListVo.RoleName) {
		e := common.ErrParam
		e.Message = "角色不存在！"
		common.SendResponse(ctx, e, roleListVo)
		return
	}

	e, _ := auth.GetE(ctx)
	for _, v := range query.RName {
		result, err := e.RemoveFilteredNamedGroupingPolicy("g", 0,
			global.UserPrefix+query.UserID.String(), global.RolePrefix+v)
		if !result {
			log.Printf("用户\"%s\"删除角色：\"%s\"失败,失败原因：\"%v\" ", query.UserID.String(), v, err)
		}
	}
	common.SendResponse(ctx, common.OK, "")
	return
}
