package model

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"one-api/common"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TokenCacheSeconds         = common.SyncFrequency
	UserId2GroupCacheSeconds  = common.SyncFrequency
	UserId2QuotaCacheSeconds  = common.SyncFrequency
	UserId2StatusCacheSeconds = common.SyncFrequency
	redisClient               = common.RDB
	ctx                       = context.Background()
)

func CacheGetTokenByKey(key string) (*Token, error) {
	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}
	var token Token
	if !common.RedisEnabled {
		err := DB.Where(keyCol+" = ?", key).First(&token).Error
		return &token, err
	}
	tokenObjectString, err := common.RedisGet(fmt.Sprintf("token:%s", key))
	if err != nil {
		err := DB.Where(keyCol+" = ?", key).First(&token).Error
		if err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		err = common.RedisSet(fmt.Sprintf("token:%s", key), string(jsonBytes), time.Duration(TokenCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set token error: " + err.Error())
		}
		return &token, nil
	}
	err = json.Unmarshal([]byte(tokenObjectString), &token)
	return &token, err
}

func CacheGetUserGroup(id int) (group string, err error) {
	if !common.RedisEnabled {
		return GetUserGroup(id)
	}
	group, err = common.RedisGet(fmt.Sprintf("user_group:%d", id))
	if err != nil {
		group, err = GetUserGroup(id)
		if err != nil {
			return "", err
		}
		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, time.Duration(UserId2GroupCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user group error: " + err.Error())
		}
	}
	return group, err
}

func CacheGetUserQuota(id int) (quota int, err error) {
	if !common.RedisEnabled {
		return GetUserQuota(id)
	}
	quotaString, err := common.RedisGet(fmt.Sprintf("user_quota:%d", id))
	if err != nil {
		quota, err = GetUserQuota(id)
		if err != nil {
			return 0, err
		}
		err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user quota error: " + err.Error())
		}
		return quota, err
	}
	quota, err = strconv.Atoi(quotaString)
	return quota, err
}

func CacheUpdateUserQuota(id int) error {
	if !common.RedisEnabled {
		return nil
	}
	quota, err := GetUserQuota(id)
	if err != nil {
		return err
	}
	err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
	return err
}

func CacheDecreaseUserQuota(id int, quota int) error {
	if !common.RedisEnabled {
		return nil
	}
	err := common.RedisDecrease(fmt.Sprintf("user_quota:%d", id), int64(quota))
	return err
}

func CacheIsUserEnabled(userId int) (bool, error) {
	if !common.RedisEnabled {
		return IsUserEnabled(userId)
	}
	enabled, err := common.RedisGet(fmt.Sprintf("user_enabled:%d", userId))
	if err == nil {
		return enabled == "1", nil
	}

	userEnabled, err := IsUserEnabled(userId)
	if err != nil {
		return false, err
	}
	enabled = "0"
	if userEnabled {
		enabled = "1"
	}
	err = common.RedisSet(fmt.Sprintf("user_enabled:%d", userId), enabled, time.Duration(UserId2StatusCacheSeconds)*time.Second)
	if err != nil {
		common.SysError("Redis set user enabled error: " + err.Error())
	}
	return userEnabled, err
}

var group2model2channels map[string]map[string][]*Channel
var channelSyncLock sync.RWMutex

func InitChannelCache() {
	var channels []*Channel
	startTime := time.Now()
	common.SysLog("start syncing channels from database")
	channels, err := GetEnableChannels()
	if err != nil {
		return
	}
	newChannelId2channel := make(map[int]*Channel)

	modelsCache := make(map[string][]string)
	groupsCache := make(map[string][]string)

	newGroup2model2channels := make(map[string]map[string][]*Channel)

	for _, channel := range channels {
		newChannelId2channel[channel.Id] = channel
		models_, ok := modelsCache[channel.Models]
		if !ok {
			models_ = strings.Split(channel.Models, ",")
			modelsCache[channel.Models] = models_
		}

		groups_, ok := groupsCache[channel.Group]
		if !ok {
			groups_ = strings.Split(channel.Group, ",")
			groupsCache[channel.Group] = groups_
		}

		for _, group := range groups_ {
			if newGroup2model2channels[group] == nil {
				newGroup2model2channels[group] = make(map[string][]*Channel)
			}
			for _, model := range models_ {
				if _, ok := newGroup2model2channels[group][model]; !ok {
					newGroup2model2channels[group][model] = make([]*Channel, 0, len(models_))
				}
				newGroup2model2channels[group][model] = append(newGroup2model2channels[group][model], channel)
			}
		}
	}

	common.SysLog("channels synced from database, took " + time.Since(startTime).String())

	// sort by priority
	for group, model2channels := range newGroup2model2channels {
		for model, channels := range model2channels {
			sort.Slice(channels, func(i, j int) bool {
				return channels[i].GetPriority() > channels[j].GetPriority()
			})
			newGroup2model2channels[group][model] = channels
		}
	}

	channelSyncLock.Lock()
	group2model2channels = newGroup2model2channels
	channelSyncLock.Unlock()
	common.SysLog("channels synced from database")
}

func SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing channels from database")
		InitChannelCache()
	}
}

func CacheGetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
	if !common.MemoryCacheEnabled {
		return GetRandomSatisfiedChannel(group, model)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()
	channels := group2model2channels[group][model]
	if len(channels) == 0 {
		return nil, errors.New("channel not found")
	}
	endIdx := len(channels)
	// choose by priority
	firstChannel := channels[0]
	if firstChannel.GetPriority() > 0 {
		for i := range channels {
			if channels[i].GetPriority() != firstChannel.GetPriority() {
				endIdx = i
				break
			}
		}
	}
	idx := rand.Intn(endIdx)
	return channels[idx], nil
}

func UpdateAllAbilities() error {
	//channelSyncLock.Lock()
	//defer channelSyncLock.Unlock()
	//channels := make([]*Channel, 0)
	//err := DB.Find(&channels).Error
	//if err != nil {
	//	return err
	//}
	//abilities := GenAbilitiesByChannels(channels)
	//err = DB.Where("1 = 1").Delete(&Ability{}).Error
	//if err != nil {
	//	return err
	//}
	//return DB.CreateInBatches(&abilities, 40).Error
	return nil
}

func GenAbilitiesByChannels(channels []*Channel) []*Ability {
	modelsCache := make(map[string][]string)
	groupsCache := make(map[string][]string)
	totalAbilities := 0
	for _, channel := range channels {
		models, ok := modelsCache[channel.Models]
		if !ok {
			models = strings.Split(channel.Models, ",")
			modelsCache[channel.Models] = models
		}

		groups, ok := groupsCache[channel.Group]
		if !ok {
			groups = strings.Split(channel.Group, ",")
			groupsCache[channel.Group] = groups
		}
		totalAbilities += len(models) * len(groups)
	}

	abilities := make([]*Ability, totalAbilities)
	var index int64 = -1
	for _, channel := range channels {
		models_, modelsOk := modelsCache[channel.Models]
		groups_, groupsOk := groupsCache[channel.Group]
		if !modelsOk || !groupsOk {
			// 这里可以添加错误处理逻辑
			continue
		}
		for _, model := range models_ {
			for _, group := range groups_ {
				ability := &Ability{
					Group:     group,
					Model:     model,
					ChannelId: channel.Id,
					Enabled:   channel.Status == common.ChannelStatusEnabled,
					Priority:  channel.Priority,
				}
				idx := atomic.AddInt64(&index, 1)
				abilities[idx] = ability
			}
		}
	}
	return abilities
}

func GenAbilitiesByChannelsWithChan(channels []*Channel) chan *Ability {
	// 返回 chan 的方式，可以避免内存占用过大
	abilityCh := make(chan *Ability, 200)
	go func(channels []*Channel) {
		defer close(abilityCh)
		modelsCache := make(map[string][]string)
		groupsCache := make(map[string][]string)
		for _, channel := range channels {
			models_, ok := modelsCache[channel.Models]
			if !ok {
				models_ = strings.Split(channel.Models, ",")
				modelsCache[channel.Models] = models_
			}

			groups_, ok := groupsCache[channel.Group]
			if !ok {
				groups_ = strings.Split(channel.Group, ",")
				groupsCache[channel.Group] = groups_
			}
			for _, model := range models_ {
				for _, group := range groups_ {
					ability := &Ability{
						Group:     group,
						Model:     model,
						ChannelId: channel.Id,
						Enabled:   channel.Status == common.ChannelStatusEnabled,
						Priority:  channel.Priority,
					}
					abilityCh <- ability
				}
			}
		}
	}(channels)
	return abilityCh
}
