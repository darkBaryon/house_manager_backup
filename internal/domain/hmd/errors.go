package hmd

import (
	"fmt"
	"strings"

	"house-manager/pkg/errcode"
)

func invalidParamf(format string, args ...any) error {
	return errcode.InvalidParam.WithError(fmt.Errorf(format, args...))
}

func notFoundf(format string, args ...any) error {
	return errcode.NotFound.WithError(fmt.Errorf(format, args...))
}

func alreadyExistsf(format string, args ...any) error {
	return errcode.AlreadyExists.WithError(fmt.Errorf(format, args...))
}

func databasef(format string, args ...any) error {
	return errcode.DatabaseError.WithError(fmt.Errorf(format, args...))
}

func mutationError(action string, err error) error {
	if err == nil {
		return nil
	}
	if errcode.FromError(err) != nil {
		return err
	}
	if duplicateErr := duplicateKeyMutationError(action, err); duplicateErr != nil {
		return duplicateErr
	}

	wrapped := fmt.Errorf("%s: %w", action, err)
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "required") ||
		strings.Contains(message, "invalid") ||
		strings.Contains(message, "not allowed") ||
		strings.Contains(message, "must be") ||
		strings.Contains(message, "is nil") ||
		strings.Contains(message, "不能为空") ||
		strings.Contains(message, "不合法") ||
		strings.Contains(message, "不能小于") {
		return errcode.InvalidParam.WithError(wrapped)
	}
	return errcode.DatabaseError.WithError(wrapped)
}

func duplicateKeyMutationError(action string, err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if !strings.Contains(message, "E11000 duplicate key error") {
		return nil
	}

	switch {
	case strings.Contains(message, "project_id_1_building_name_1_active_unique"):
		return alreadyExistsf("当前项目下已存在同名楼栋")
	case strings.Contains(message, "project_id_1_room_type_name_1_active_unique"):
		return alreadyExistsf("当前项目下已存在同名房型")
	case strings.Contains(message, "building_id_1_room_type_name_1_active_unique"):
		return alreadyExistsf("当前楼栋下已存在同名房型")
	case strings.Contains(message, "building_id_1_room_no_1_active_unique"):
		return alreadyExistsf("当前楼栋下已存在相同房间号")
	case strings.Contains(message, "decentralized_id_1_room_no_1_active_unique"):
		return alreadyExistsf("当前小区下已存在相同房间号")
	case strings.Contains(message, "city_1_district_1_community_name_1_active_unique"):
		return alreadyExistsf("当前城市和区域下已存在同名小区")
	default:
		return alreadyExistsf("存在重复数据，请检查是否重复创建")
	}
}
