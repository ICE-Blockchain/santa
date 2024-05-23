// SPDX-License-Identifier: ice License 1.0

package badges

import (
	"context"
	_ "embed"
	"io"

	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
)

// Public API.

const (
	LevelGroupType  GroupType = "level"
	CoinGroupType   GroupType = "coin"
	SocialGroupType GroupType = "social"
)

//nolint:gochecknoglobals // .
var (
	ErrRelationNotFound = storage.ErrRelationNotFound
	ErrHidden           = errors.New("badges are hidden")
	ErrRaceCondition    = errors.New("race condition")

	AllTypes             []Type
	LevelTypeOrder       map[Type]int
	CoinTypeOrder        map[Type]int
	SocialTypeOrder      map[Type]int
	AllTypeOrder         map[Type]int
	LevelTypeNames       map[Type]string
	CoinTypeNames        map[Type]string
	SocialTypeNames      map[Type]string
	GroupTypeForEachType map[Type]GroupType
	AllNames             map[GroupType]map[Type]string
	AllGroups            map[GroupType][]Type
	GroupsOrderSummaries = [3]GroupType{
		SocialGroupType,
		CoinGroupType,
		LevelGroupType,
	}
	Milestones map[Type]AchievingRange
)

type (
	Type           string
	GroupType      string
	AchievingRange struct {
		Name          string `json:"-"`
		FromInclusive uint64 `json:"fromInclusive,omitempty"`
		ToInclusive   uint64 `json:"toInclusive,omitempty"`
	}
	Badge struct {
		Name                        string         `json:"name"`
		Type                        Type           `json:"-"`
		GroupType                   GroupType      `json:"type"` //nolint:tagliatelle // Intended.
		AchievingRange              AchievingRange `json:"achievingRange"`
		PercentageOfUsersInProgress float64        `json:"percentageOfUsersInProgress"`
		Achieved                    bool           `json:"achieved"`
	}
	BadgeSummary struct {
		Name      string    `json:"name"`
		GroupType GroupType `json:"type"` //nolint:tagliatelle // Intended.
		Index     uint64    `json:"index"`
		LastIndex uint64    `json:"lastIndex"`
	}
	AchievedBadge struct {
		UserID         string    `json:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		Type           Type      `json:"type" example:"c1"`
		Name           string    `json:"name" example:"Glacial Polly"`
		GroupType      GroupType `json:"groupType" example:"coin"`
		AchievedBadges uint64    `json:"achievedBadges,omitempty" example:"3"`
	}
	ReadRepository interface {
		GetBadges(ctx context.Context, groupType GroupType, userID string) ([]*Badge, error)
		GetSummary(ctx context.Context, userID string) ([]*BadgeSummary, error)
	}
	WriteRepository interface{} //nolint:revive // .
	Repository      interface {
		io.Closer

		ReadRepository
		WriteRepository
	}
	Processor interface {
		Repository
		CheckHealth(ctx context.Context) error
	}
)

// Private API.

const (
	applicationYamlKey          = "badges"
	requestingUserIDCtxValueKey = "requestingUserIDCtxValueKey"
	percent100                  = 100.0
)

// .
var (
	//go:embed DDL.sql
	ddl string
)

type (
	progress struct {
		AchievedBadges  *users.Enum[Type] `json:"achievedBadges,omitempty" example:"c1,l1,l2,c2"`
		UserID          string            `json:"userId,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		FriendsInvited  int64             `json:"friendsInvited,omitempty" example:"3"`
		CompletedLevels int64             `json:"completedLevels,omitempty" example:"3"`
		Balance         int64             `json:"balance,omitempty" example:"1232323232"`
		HideBadges      bool              `json:"hideBadges,omitempty" example:"false"`
	}
	statistics struct {
		Type       Type      `db:"badge_type"`
		GroupType  GroupType `db:"badge_group_type"`
		AchievedBy uint64
	}
	tryAchievedBadgesCommandSource struct {
		*processor
	}
	achievedBadgesSource struct {
		*processor
	}
	userTableSource struct {
		*processor
	}

	friendsInvitedSource struct {
		*processor
	}

	completedLevelsSource struct {
		*processor
	}
	balancesTableSource struct {
		*processor
	}
	globalTableSource struct {
		*processor
	}
	repository struct {
		cfg      *config
		shutdown func() error
		db       *storage.DB
		mb       messagebroker.Client
	}
	processor struct {
		*repository
	}
	config struct {
		Levels               []*AchievingRange        `yaml:"levels"`
		Coins                []*AchievingRange        `yaml:"coins"`
		Socials              []*AchievingRange        `yaml:"socials"`
		messagebroker.Config `mapstructure:",squash"` //nolint:tagliatelle // Nope.
	}
)
