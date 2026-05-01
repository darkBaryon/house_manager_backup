package service

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"
	"house-manager/pkg/cache"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RoomFinder 房间数据查询接口（service 消费 repo）
type RoomFinder interface {
	FindById(ctx context.Context, id bson.ObjectID) (*model.Room, error)
	FindList(ctx context.Context, filter bson.M, req model.PageReq) ([]model.Room, int64, error)
}

// RoomService 房间业务层
type RoomService struct {
	Repo  RoomFinder
	Cache *cache.Cache
}

// NewRoomService 创建房间业务层实例
func NewRoomService(repo RoomFinder, c *cache.Cache) *RoomService {
	return &RoomService{Repo: repo, Cache: c}
}

// GetDetail 查询房间明细（Cache-Aside）
func (s *RoomService) GetDetail(ctx context.Context, idHex string) (*model.Room, error) {
	objId, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, errcode.RoomInvalidId
	}

	cacheKey := "room:detail:" + idHex
	return cache.GetSet(s.Cache, ctx, cacheKey, 5*time.Minute, func(ctx context.Context) (*model.Room, error) {
		room, err := s.Repo.FindById(ctx, objId)
		if err != nil {
			return nil, errcode.DatabaseError.WithError(err)
		}
		if room == nil {
			return nil, errcode.RoomNotFound
		}
		return room, nil
	})
}

// roomListResult 缓存列表结果
type roomListResult struct {
	Rooms []model.Room `json:"rooms"`
	Total int64        `json:"total"`
}

// GetList 查询房间列表（分页 + 缓存）
func (s *RoomService) GetList(ctx context.Context, req model.PageReq) ([]model.Room, int64, error) {
	cacheKey := fmt.Sprintf("room:list:%d:%d", req.Offset, req.Limit)
	result, err := cache.GetSet(s.Cache, ctx, cacheKey, 5*time.Minute, func(ctx context.Context) (roomListResult, error) {
		rooms, total, err := s.Repo.FindList(ctx, bson.M{}, req)
		if err != nil {
			return roomListResult{}, errcode.DatabaseError.WithError(err)
		}
		return roomListResult{Rooms: rooms, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return result.Rooms, result.Total, nil
}
