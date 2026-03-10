package model

import (
	"errors"
	"fmt"
)

// ============ 状态流转错误定义 ============

var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceAlreadyExists  = errors.New("resource already exists")
	ErrFileNotOnDisk          = errors.New("file not found on disk")
	ErrDatabaseSyncFailed     = errors.New("database sync failed")
)

// ============ 投稿状态流转 ============

// SubmissionTransition 定义投稿状态流转规则
// pending -> approved (审核通过)
// pending -> rejected (审核驳回)
// rejected -> pending (重新提交，可选)
var SubmissionTransitions = map[string][]string{
	StatusPending:  {StatusApproved, StatusRejected},
	StatusApproved: {}, // 已通过不可变更
	StatusRejected: {}, // 已驳回不可变更（如需重新提交，建议新建记录）
}

// CanTransitionSubmission 检查投稿状态是否可以转换
func CanTransitionSubmission(from, to string) bool {
	allowedStates, exists := SubmissionTransitions[from]
	if !exists {
		return false
	}
	for _, s := range allowedStates {
		if s == to {
			return true
		}
	}
	return false
}

// ValidateSubmissionTransition 验证并返回错误
func ValidateSubmissionTransition(from, to string) error {
	if !CanTransitionSubmission(from, to) {
		return fmt.Errorf("%w: submission cannot transition from %s to %s", ErrInvalidStateTransition, from, to)
	}
	return nil
}

// ============ 资源状态流转 ============

// ResourceTransition 定义资源（文件/文件夹）状态流转规则
// active -> offline (下架)
// active -> deleted (删除)
// offline -> active (恢复上架)
// offline -> deleted (删除)
// deleted -> active (恢复，从回收站恢复)
var ResourceTransitions = map[string][]string{
	ResourceActive:  {ResourceOffline, ResourceDeleted},
	ResourceOffline: {ResourceActive, ResourceDeleted},
	ResourceDeleted: {ResourceActive}, // 支持从回收站恢复
}

// CanTransitionResource 检查资源状态是否可以转换
func CanTransitionResource(from, to string) bool {
	allowedStates, exists := ResourceTransitions[from]
	if !exists {
		return false
	}
	for _, s := range allowedStates {
		if s == to {
			return true
		}
	}
	return false
}

// ValidateResourceTransition 验证资源状态转换
func ValidateResourceTransition(from, to string) error {
	if !CanTransitionResource(from, to) {
		return fmt.Errorf("%w: resource cannot transition from %s to %s", ErrInvalidStateTransition, from, to)
	}
	return nil
}

// ============ 举报状态流转 ============

// ReportTransitions 定义举报状态流转规则
var ReportTransitions = map[string][]string{
	StatusPending:  {StatusApproved, StatusRejected},
	StatusApproved: {}, // 已处理不可变更
	StatusRejected: {}, // 已驳回不可变更
}

// CanTransitionReport 检查举报状态是否可以转换
func CanTransitionReport(from, to string) bool {
	allowedStates, exists := ReportTransitions[from]
	if !exists {
		return false
	}
	for _, s := range allowedStates {
		if s == to {
			return true
		}
	}
	return false
}

// ============ 状态同步规则 ============

// SyncAction 定义同步动作
type SyncAction string

const (
	SyncActionMoveToRepository SyncAction = "move_to_repository" // staging -> repository
	SyncActionMoveToTrash      SyncAction = "move_to_trash"      // repository -> trash
	SyncActionRestoreFromTrash SyncAction = "restore_from_trash" // trash -> repository
	SyncActionDeletePermanent  SyncAction = "delete_permanent"   // 物理删除
	SyncActionNone             SyncAction = "none"               // 无需同步
)

// StateChangeResult 状态变更结果
type StateChangeResult struct {
	OldState   string
	NewState   string
	SyncAction SyncAction
	SourcePath string // 源路径
	TargetPath string // 目标路径
}

// GetSubmissionApprovalSync 获取投稿审核通过时的同步动作
func GetSubmissionApprovalSync(stagingPath, repositoryPath string) StateChangeResult {
	return StateChangeResult{
		OldState:   StatusPending,
		NewState:   StatusApproved,
		SyncAction: SyncActionMoveToRepository,
		SourcePath: stagingPath,
		TargetPath: repositoryPath,
	}
}

// GetResourceOfflineSync 获取资源下架时的同步动作
func GetResourceOfflineSync(repositoryPath, trashPath string) StateChangeResult {
	return StateChangeResult{
		OldState:   ResourceActive,
		NewState:   ResourceOffline,
		SyncAction: SyncActionMoveToTrash,
		SourcePath: repositoryPath,
		TargetPath: trashPath,
	}
}

// GetResourceRestoreSync 获取资源恢复时的同步动作
func GetResourceRestoreSync(trashPath, repositoryPath string) StateChangeResult {
	return StateChangeResult{
		OldState:   ResourceOffline,
		NewState:   ResourceActive,
		SyncAction: SyncActionRestoreFromTrash,
		SourcePath: trashPath,
		TargetPath: repositoryPath,
	}
}

// ============ 状态显示文本 ============

// SubmissionStatusText 投稿状态显示文本
var SubmissionStatusText = map[string]string{
	StatusPending:  "待审核",
	StatusApproved: "已通过",
	StatusRejected: "已驳回",
}

// ResourceStatusText 资源状态显示文本
var ResourceStatusText = map[string]string{
	ResourceActive:  "正常",
	ResourceOffline: "已下架",
	ResourceDeleted: "已删除",
}

// GetSubmissionStatusText 获取投稿状态显示文本
func GetSubmissionStatusText(status string) string {
	if text, ok := SubmissionStatusText[status]; ok {
		return text
	}
	return status
}

// GetResourceStatusText 获取资源状态显示文本
func GetResourceStatusText(status string) string {
	if text, ok := ResourceStatusText[status]; ok {
		return text
	}
	return status
}

// ============ 状态校验辅助 ============

// IsValidSubmissionStatus 检查是否是有效的投稿状态
func IsValidSubmissionStatus(status string) bool {
	switch status {
	case StatusPending, StatusApproved, StatusRejected:
		return true
	}
	return false
}

// IsValidResourceStatus 检查是否是有效的资源状态
func IsValidResourceStatus(status string) bool {
	switch status {
	case ResourceActive, ResourceOffline, ResourceDeleted:
		return true
	}
	return false
}

// IsPublicResource 检查资源是否对外可见
func IsPublicResource(status string) bool {
	return status == ResourceActive
}
