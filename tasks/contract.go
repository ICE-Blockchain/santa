// SPDX-License-Identifier: ice License 1.0

package tasks

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
)

// Public API.

const (
	ClaimUsernameType        Type = "claim_username"
	StartMiningType          Type = "start_mining"
	UploadProfilePictureType Type = "upload_profile_picture"
	FollowUsOnTwitterType    Type = "follow_us_on_twitter"
	JoinTelegramType         Type = "join_telegram"
	InviteFriendsType        Type = "invite_friends"
)

var (
	ErrRelationNotFound = storage.ErrRelationNotFound
	ErrRaceCondition    = errors.New("race condition")
)

type (
	Type string
	Data struct {
		TwitterUserHandle  string `json:"twitterUserHandle,omitempty" example:"jdoe2"`
		TelegramUserHandle string `json:"telegramUserHandle,omitempty" example:"jdoe1"`
		VerificationCode   string `json:"verificationCode,omitempty" example:"ABC"`
		RequiredQuantity   uint64 `json:"requiredQuantity,omitempty" example:"3"`
	}
	Metadata struct {
		Title            string `json:"title,omitempty" example:"Claim username"`
		ShortDescription string `json:"shortDescription,omitempty" example:"Short description"`
		LongDescription  string `json:"longDescription,omitempty" example:"Long description"`
		IconURL          string `json:"iconUrl,omitempty" example:"https://app.ice.com/web/invite.svg"`
	}
	Task struct {
		Data      *Data     `json:"data,omitempty"`
		Metadata  *Metadata `json:"metadata,omitempty"`
		UserID    string    `json:"userId,omitempty" swaggerignore:"true" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		Type      Type      `json:"type" example:"claim_username"`
		Prize     float64   `json:"prize" example:"200.0"`
		Completed bool      `json:"completed" example:"false"`
	}
	CompletedTask struct {
		UserID         string  `json:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		Type           Type    `json:"type" example:"claim_username"`
		CompletedTasks uint64  `json:"completedTasks,omitempty" example:"3"`
		Prize          float64 `json:"prize,omitempty" example:"200"`
	}
	ReadRepository interface {
		GetTasks(ctx context.Context, userID, languageCode string) ([]*Task, error)
	}
	WriteRepository interface {
		PseudoCompleteTask(ctx context.Context, task *Task) error
	}
	Repository interface {
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
	applicationYamlKey = "tasks"
	defaultLanguage    = "en"
)

// .
var (
	//go:embed DDL.sql
	ddl string

	//nolint:gochecknoglobals // Its loaded once at startup.
	allTaskTemplates map[Type]map[languageCode]*taskTemplate

	//go:embed translations
	translations embed.FS
)

type (
	languageCode = string
	progress     struct {
		CompletedTasks       *users.Enum[Type] `json:"completedTasks,omitempty" example:"claim_username,start_mining"`
		PseudoCompletedTasks *users.Enum[Type] `json:"pseudoCompletedTasks,omitempty" example:"claim_username,start_mining"`
		TwitterUserHandle    *string           `json:"twitterUserHandle,omitempty" example:"jdoe2"`
		TelegramUserHandle   *string           `json:"telegramUserHandle,omitempty" example:"jdoe1"`
		UserID               string            `json:"userId,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		FriendsInvited       uint64            `json:"friendsInvited,omitempty" example:"3"`
		UsernameSet          bool              `json:"usernameSet,omitempty" example:"true"`
		ProfilePictureSet    bool              `json:"profilePictureSet,omitempty" example:"true"`
		MiningStarted        bool              `json:"miningStarted,omitempty" example:"true"`
	}
	taskTemplate struct {
		title, shortDescription, longDescription *template.Template
		Title                                    string `json:"title"`            //nolint:revive // That's intended.
		ShortDescription                         string `json:"shortDescription"` //nolint:revive // That's intended.
		LongDescription                          string `json:"longDescription"`  //nolint:revive // That's intended.
	}
	tryCompleteTasksCommandSource struct {
		*processor
	}
	miningSessionSource struct {
		*processor
	}
	userTableSource struct {
		*processor
	}

	friendsInvitedSource struct {
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
		TenantName string `yaml:"tenantName" mapstructure:"tenantName"`
		TasksList  []struct {
			Type  string  `yaml:"type" mapstructure:"type"`
			Icon  string  `yaml:"icon" mapstructure:"icon"`
			Prize float64 `yaml:"prize" mapstructure:"prize"`
		} `yaml:"tasksList" mapstructure:"tasksList"`
		messagebroker.Config   `mapstructure:",squash"` //nolint:tagliatelle // Nope.
		RequiredFriendsInvited uint64                   `yaml:"requiredFriendsInvited"`
	}
)
