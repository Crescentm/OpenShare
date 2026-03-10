package model

// PermissionInfo 权限项信息
type PermissionInfo struct {
	Code        string `json:"code"`        // 权限代码
	Name        string `json:"name"`        // 显示名称
	Description string `json:"description"` // 描述
	Group       string `json:"group"`       // 分组
}

// PermissionGroup 权限分组
type PermissionGroup struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Permissions []PermissionInfo `json:"permissions"`
}

// 权限分组常量
const (
	PermGroupContent = "content" // 内容管理
	PermGroupAudit   = "audit"   // 审核管理
	PermGroupSystem  = "system"  // 系统管理
)

// PermissionRegistry 权限注册表
// 提供权限元数据，便于前端展示和权限管理
var PermissionRegistry = map[string]PermissionInfo{
	PermissionReviewSubmission: {
		Code:        PermissionReviewSubmission,
		Name:        "审核资料",
		Description: "审核用户上传的资料，可通过或驳回",
		Group:       PermGroupAudit,
	},
	PermissionPublishAnnounce: {
		Code:        PermissionPublishAnnounce,
		Name:        "发布公告",
		Description: "发布、编辑和删除系统公告",
		Group:       PermGroupSystem,
	},
	PermissionEditFile: {
		Code:        PermissionEditFile,
		Name:        "编辑资料",
		Description: "修改资料标题、描述、Tag等信息",
		Group:       PermGroupContent,
	},
	PermissionDeleteFile: {
		Code:        PermissionDeleteFile,
		Name:        "删除资料",
		Description: "删除或下架资料",
		Group:       PermGroupContent,
	},
	PermissionManageTag: {
		Code:        PermissionManageTag,
		Name:        "管理标签",
		Description: "创建、编辑、删除和合并标签",
		Group:       PermGroupContent,
	},
	PermissionManageReport: {
		Code:        PermissionManageReport,
		Name:        "处理举报",
		Description: "查看和处理用户举报",
		Group:       PermGroupAudit,
	},
	PermissionViewLog: {
		Code:        PermissionViewLog,
		Name:        "查看日志",
		Description: "查看系统操作日志",
		Group:       PermGroupSystem,
	},
}

// GetPermissionGroups 获取按分组组织的权限列表
func GetPermissionGroups() []PermissionGroup {
	groups := map[string]*PermissionGroup{
		PermGroupContent: {
			Code: PermGroupContent,
			Name: "内容管理",
		},
		PermGroupAudit: {
			Code: PermGroupAudit,
			Name: "审核管理",
		},
		PermGroupSystem: {
			Code: PermGroupSystem,
			Name: "系统管理",
		},
	}

	// 按分组整理权限
	for _, perm := range PermissionRegistry {
		if group, ok := groups[perm.Group]; ok {
			group.Permissions = append(group.Permissions, perm)
		}
	}

	// 转换为切片
	result := make([]PermissionGroup, 0, len(groups))
	// 按固定顺序返回
	for _, code := range []string{PermGroupContent, PermGroupAudit, PermGroupSystem} {
		if group := groups[code]; group != nil && len(group.Permissions) > 0 {
			result = append(result, *group)
		}
	}

	return result
}

// GetPermissionInfo 获取单个权限的信息
func GetPermissionInfo(code string) (PermissionInfo, bool) {
	info, ok := PermissionRegistry[code]
	return info, ok
}

// IsValidPermission 检查权限代码是否有效
func IsValidPermission(code string) bool {
	_, ok := PermissionRegistry[code]
	return ok
}

// ValidatePermissions 验证权限列表，返回无效的权限代码
func ValidatePermissions(codes []string) []string {
	invalid := make([]string, 0)
	for _, code := range codes {
		if !IsValidPermission(code) {
			invalid = append(invalid, code)
		}
	}
	return invalid
}
