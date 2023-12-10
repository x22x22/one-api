package model

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model     string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool   `json:"enabled"`
	Priority  *int64 `json:"priority" gorm:"bigint;default:0;index"`
}

func GetGroupModels(group string) []string {
	var models []string
	var channels []*Channel
	channels, err := GetEnableChannels()
	if err != nil {
		return models
	}
	abilities := GenAbilitiesByChannelsWithChan(channels)
	// 使用map来确保models的唯一性
	modelsMap := make(map[string]struct{})
	for ability := range abilities {
		if ability.Group == group && ability.Enabled {
			modelsMap[ability.Model] = struct{}{}
		}
	}
	// 将map的键转换为slice
	for model := range modelsMap {
		models = append(models, model)
	}
	return models
}

func GetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
	var satisfiedAbilities []*Ability
	channels, err := GetEnableChannels()
	if err != nil {
		return nil, err
	}
	abilities := GenAbilitiesByChannelsWithChan(channels)
	for ability := range abilities {
		if ability.Group == group && ability.Model == model && ability.Enabled {
			satisfiedAbilities = append(satisfiedAbilities, ability)
		}
	}
	if len(satisfiedAbilities) == 0 {
		return nil, errors.New("no satisfied abilities found")
	}
	// 获取最大优先级
	maxPriority := int64(math.MinInt64)
	for _, ability := range satisfiedAbilities {
		if ability.Priority != nil && *ability.Priority > maxPriority {
			maxPriority = *ability.Priority
		}
	}
	// 过滤出最大优先级的abilities
	var maxPriorityAbilities []*Ability
	for _, ability := range satisfiedAbilities {
		if ability.Priority != nil && *ability.Priority == maxPriority {
			maxPriorityAbilities = append(maxPriorityAbilities, ability)
		}
	}
	// 随机选择一个
	rand.NewSource(time.Now().UnixNano())
	randomIndex := rand.Intn(len(maxPriorityAbilities))
	selectedAbility := maxPriorityAbilities[randomIndex]
	// 根据ChannelId查找对应的Channel
	channel := Channel{}
	channel.Id = selectedAbility.ChannelId
	for _, c := range channels { // 假设channels是全局可用的
		if c.Id == selectedAbility.ChannelId {
			channel = *c
			break
		}
	}
	if channel.Id == 0 {
		return nil, errors.New("channel not found")
	}
	return &channel, nil
}

//func GetGroupModels(group string) []string {
//	var models []string
//	// Find distinct models
//	groupCol := "`group`"
//	if common.UsingPostgreSQL {
//		groupCol = `"group"`
//	}
//	DB.Table("abilities").Where(groupCol+" = ? and enabled = ?", group, true).Distinct("model").Pluck("model", &models)
//	return models
//}

//func GetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
//	ability := Ability{}
//	groupCol := "`group`"
//	trueVal := "1"
//	if common.UsingPostgreSQL {
//		groupCol = `"group"`
//		trueVal = "true"
//	}
//
//	var err error = nil
//	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
//	channelQuery := DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
//	if common.UsingSQLite || common.UsingPostgreSQL {
//		err = channelQuery.Order("RANDOM()").First(&ability).Error
//	} else {
//		err = channelQuery.Order("RAND()").First(&ability).Error
//	}
//	if err != nil {
//		return nil, err
//	}
//	channel := Channel{}
//	channel.Id = ability.ChannelId
//	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
//	return &channel, err
//}

func (channel *Channel) AddAbilities() error {
	return nil
	//models_ := strings.Split(channel.Models, ",")
	//groups_ := strings.Split(channel.Group, ",")
	//abilities := make([]Ability, 0, len(models_))
	//for _, model := range models_ {
	//	for _, group := range groups_ {
	//		ability := Ability{
	//			Group:     group,
	//			Model:     model,
	//			ChannelId: channel.Id,
	//			Enabled:   channel.Status == common.ChannelStatusEnabled,
	//			Priority:  channel.Priority,
	//		}
	//		abilities = append(abilities, ability)
	//	}
	//}
	//return DB.Create(&abilities).Error
}

func (channel *Channel) DeleteAbilities() error {
	return nil
	//return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	return nil
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	//err := channel.DeleteAbilities()
	//if err != nil {
	//	return err
	//}
	//// Then add new abilities
	//err = channel.AddAbilities()
	//if err != nil {
	//	return err
	//}
	//return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return nil
	//return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}
