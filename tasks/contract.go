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
	ClaimUsernameType Type = "claim_username"
	StartMiningType   Type = "start_mining"

	// V1.
	UploadProfilePictureType Type = "upload_profile_picture"
	FollowUsOnTwitterType    Type = "follow_us_on_twitter"
	JoinTwitterType          Type = "join_twitter"
	JoinTelegramType         Type = "join_telegram"
	InviteFriendsType        Type = "invite_friends"

	// V2.
	JoinTwitterUsType          Type = "join_twitter"
	JoinTelegramUsType         Type = "join_telegram"
	JoinRedditIONType          Type = "join_reddit_ion"
	JoinInstagramIONType       Type = "join_instagram_ion"
	JoinTwitterIONType         Type = "join_twitter_ion"
	JoinTelegramIONType        Type = "join_telegram_ion"
	JoinYoutubeType            Type = "join_youtube"
	JoinTwitterMultiversxType  Type = "join_twitter_multiversx"
	JoinTwitterXPortalType     Type = "join_twitter_xportal"
	JoinTelegramMultiversxType Type = "join_telegram_multiversx"
	JoinBullishCMCType         Type = "join_bullish_cmc"
	JoinIONCMCType             Type = "join_ion_cmc"
	JoinWatchListCMCType       Type = "join_watchlist_cmc"
	JoinPortfolioCoinGeckoType Type = "join_portfolio_coingecko"
	JoinHoldCoinType           Type = "join_holdcoin"
	JoinHumanType              Type = "join_human"
	JoinHipoType               Type = "join_hipo"
	JoinFreedogsType           Type = "join_freedogs"
	JoinAtheneType             Type = "join_athene"
	JoinKoloType               Type = "join_kolo"
	JoinDucksType              Type = "join_ducks"
	JoinCMCTONType             Type = "join_cmc_ton"
	JoinCMCSSOLType            Type = "join_cmc_sol"
	JoinCMCBNBType             Type = "join_cmc_bnb"
	JoinCMCETHType             Type = "join_cmc_eth"
	JoinCMCBTCType             Type = "join_cmc_btc"
	JoinCMCPNUType             Type = "join_cmc_pnut"
	JoinCMCADAType             Type = "join_cmc_ada"
	JoinCMCDOGEType            Type = "join_cmc_doge"
	JoinCMCXRPType             Type = "join_cmc_xrp"
	JoinCMCACTType             Type = "join_cmc_act"
	JoinBearfiType             Type = "join_bearfi"
	JoinBoinkersType           Type = "join_boinkers"
	JoinDejenDogType           Type = "join_dejendog"
	JoinCatGoldMinerType       Type = "join_catgoldminer"
	JoinTonKombatType          Type = "join_tonkombat"
	JoinTonAIType              Type = "join_tonai"
	JoinPigsType               Type = "join_pigs"
	JoinCapybaraType           Type = "join_capybara"
	JoinSidekickType           Type = "join_sidekick"
	JoinIcebergType            Type = "join_iceberg"
	JoinGoatsType              Type = "join_goats"
	JoinTapcoinsType           Type = "join_tapcoins"
	JoinTokyobeastType         Type = "join_tokyobeast"
	JoinTwitterPichainType     Type = "join_twitter_pichain"
	JoinSugarType              Type = "join_sugar"

	WatchVideoWithCodeConfirmation1Type Type = "watch_video_with_code_confirmation_1"
	InviteFriends5Type                  Type = "invite_friends_5"
	InviteFriends10Type                 Type = "invite_friends_10"
	InviteFriends25Type                 Type = "invite_friends_25"
	InviteFriends50Type                 Type = "invite_friends_50"
	InviteFriends100Type                Type = "invite_friends_100"
	InviteFriends200Type                Type = "invite_friends_200"

	SignUpCallfluentType   Type = "signup_callfluent"
	SignUpSaucesType       Type = "signup_sauces"
	SignUpSealsendType     Type = "signup_sealsend"
	SignUpSunwavesType     Type = "signup_sunwaves"
	SignUpDoctorxType      Type = "signup_doctorx"
	SignUpCryptoMayorsType Type = "signup_cryptomayors"

	ClaimLevelBadge1Type Type = "claim_badge_l1"
	ClaimLevelBadge2Type Type = "claim_badge_l2"
	ClaimLevelBadge3Type Type = "claim_badge_l3"
	ClaimLevelBadge4Type Type = "claim_badge_l4"
	ClaimLevelBadge5Type Type = "claim_badge_l5"
	ClaimLevelBadge6Type Type = "claim_badge_l6"

	ClaimCoinBadge1Type  Type = "claim_badge_c1"
	ClaimCoinBadge2Type  Type = "claim_badge_c2"
	ClaimCoinBadge3Type  Type = "claim_badge_c3"
	ClaimCoinBadge4Type  Type = "claim_badge_c4"
	ClaimCoinBadge5Type  Type = "claim_badge_c5"
	ClaimCoinBadge6Type  Type = "claim_badge_c6"
	ClaimCoinBadge7Type  Type = "claim_badge_c7"
	ClaimCoinBadge8Type  Type = "claim_badge_c8"
	ClaimCoinBadge9Type  Type = "claim_badge_c9"
	ClaimCoinBadge10Type Type = "claim_badge_c10"

	ClaimSocialBadge1Type  Type = "claim_badge_s1"
	ClaimSocialBadge2Type  Type = "claim_badge_s2"
	ClaimSocialBadge3Type  Type = "claim_badge_s3"
	ClaimSocialBadge4Type  Type = "claim_badge_s4"
	ClaimSocialBadge5Type  Type = "claim_badge_s5"
	ClaimSocialBadge6Type  Type = "claim_badge_s6"
	ClaimSocialBadge7Type  Type = "claim_badge_s7"
	ClaimSocialBadge8Type  Type = "claim_badge_s8"
	ClaimSocialBadge9Type  Type = "claim_badge_s9"
	ClaimSocialBadge10Type Type = "claim_badge_s10"

	ClaimLevel1Type    Type = "claim_level_1"
	ClaimLevel2Type    Type = "claim_level_2"
	ClaimLevel3Type    Type = "claim_level_3"
	ClaimLevel4Type    Type = "claim_level_4"
	ClaimLevel5Type    Type = "claim_level_5"
	ClaimLevel6Type    Type = "claim_level_6"
	ClaimLevel7Type    Type = "claim_level_7"
	ClaimLevel8Type    Type = "claim_level_8"
	ClaimLevel9Type    Type = "claim_level_9"
	ClaimLevel10Type   Type = "claim_level_10"
	ClaimLevel11Type   Type = "claim_level_11"
	ClaimLevel12Type   Type = "claim_level_12"
	ClaimLevel13Type   Type = "claim_level_13"
	ClaimLevel14Type   Type = "claim_level_14"
	ClaimLevel15Type   Type = "claim_level_15"
	ClaimLevel16Type   Type = "claim_level_16"
	ClaimLevel17Type   Type = "claim_level_17"
	ClaimLevel18Type   Type = "claim_level_18"
	ClaimLevel19Type   Type = "claim_level_19"
	ClaimLevel20Type   Type = "claim_level_20"
	ClaimLevel21Type   Type = "claim_level_21"
	MiningStreak7Type  Type = "mining_streak_7"
	MiningStreak14Type Type = "mining_streak_14"
	MiningStreak30Type Type = "mining_streak_30"

	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusPending   TaskStatus = "pending"

	TaskGroupBadgeSocial   = "claim_badge_social"
	TaskGroupBadgeCoin     = "claim_badge_coin"
	TaskGroupBadgeLevel    = "claim_badge_level"
	TaskGroupLevel         = "claim_level"
	TaskGroupInviteFriends = "invite_friends"
	TaskGroupMiningStreak  = "mining_streak"
)

var (
	ErrRelationNotFound          = storage.ErrRelationNotFound
	ErrRaceCondition             = errors.New("race condition")
	ErrInvalidSocialProperties   = errors.New("wrong social handle")
	ErrWrongRequestedTasksStatus = errors.New("wrong requested tasks status")
	ErrNotFound                  = errors.New("not found")
	ErrNotSupported              = errors.New("not supported")
	ErrTaskNotCompleted          = errors.New("task not completed")

	//nolint:gochecknoglobals // It's just for more descriptive validation messages.
	AllTypes = [6]Type{
		ClaimUsernameType,
		StartMiningType,
		UploadProfilePictureType,
		FollowUsOnTwitterType,
		JoinTelegramType,
		InviteFriendsType,
	}
	//nolint:gochecknoglobals // It's just for more descriptive validation messages.
	TypeOrder = map[Type]int{
		ClaimUsernameType:        0,
		StartMiningType:          1,
		UploadProfilePictureType: 2,
		FollowUsOnTwitterType:    3,
		JoinTelegramType:         4,
		InviteFriendsType:        5,
	}
)

type (
	Type       string
	TaskStatus string
	Data       struct {
		TwitterUserHandle  string `json:"twitterUserHandle,omitempty" example:"jdoe2"`
		TelegramUserHandle string `json:"telegramUserHandle,omitempty" example:"jdoe1"`
		VerificationCode   string `json:"verificationCode,omitempty" example:"ABC"`
		RequiredQuantity   uint64 `json:"requiredQuantity,omitempty" example:"3"`
	}
	Metadata struct {
		Title            string `json:"title,omitempty" example:"Claim username"`
		ShortDescription string `json:"shortDescription,omitempty" example:"Short description"`
		LongDescription  string `json:"longDescription,omitempty" example:"Long description"`
		ErrorDescription string `json:"errorDescription,omitempty" example:"Error description"`
		IconURL          string `json:"iconUrl,omitempty" example:"https://app.ice.com/web/invite.svg"`
		TaskURL          string `json:"taskUrl,omitempty" example:"https://x.com/ice_blockchain"`
	}
	Task struct {
		Data       *Data     `json:"data,omitempty"`
		Metadata   *Metadata `json:"metadata,omitempty"`
		UserID     string    `json:"userId,omitempty" swaggerignore:"true" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		Group      string    `json:"-" swaggerignore:"true"`
		Type       Type      `json:"type" example:"claim_username"`
		Prize      float64   `json:"prize" example:"200.0"`
		GroupIndex uint64    `json:"-" swaggerignore:"true"`
		Completed  bool      `json:"completed" example:"false"`
	}
	CompletedTask struct {
		UserID         string  `json:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		Type           Type    `json:"type" example:"claim_username"`
		CompletedTasks uint64  `json:"completedTasks,omitempty" example:"3"`
		Prize          float64 `json:"prize,omitempty" example:"200"`
	}
	ReadRepository interface {
		GetTasks(ctx context.Context, userID, languageCode string, requestedStatus TaskStatus) ([]*Task, error)
		GetTask(ctx context.Context, userID, language string, taskType Type) (resp *Task, err error)
	}
	WriteRepository interface {
		PseudoCompleteTask(ctx context.Context, task *Task, dryRun bool) error
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
		CompletedLevels      *users.Enum[string] `json:"completedLevels,omitempty" example:"1,2"`
		AchievedBadges       *users.Enum[string] `json:"achievedBadges,omitempty" example:"c1,l1,l2,c2"`
		CompletedTasks       *users.Enum[Type]   `json:"completedTasks,omitempty" example:"claim_username,start_mining"`
		PseudoCompletedTasks *users.Enum[Type]   `json:"pseudoCompletedTasks,omitempty" example:"claim_username,start_mining"`
		TwitterUserHandle    *string             `json:"twitterUserHandle,omitempty" example:"jdoe2"`
		TelegramUserHandle   *string             `json:"telegramUserHandle,omitempty" example:"jdoe1"`
		UserID               string              `json:"userId,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		FriendsInvited       uint64              `json:"friendsInvited,omitempty" example:"3"`
		MiningStreak         uint64              `json:"miningStreak,omitempty" example:"3"`
		UsernameSet          bool                `json:"usernameSet,omitempty" example:"true"`
		ProfilePictureSet    bool                `json:"profilePictureSet,omitempty" example:"true"`
		MiningStarted        bool                `json:"miningStarted,omitempty" example:"true"`
	}
	taskTemplate struct {
		title, shortDescription, longDescription, errorDescription *template.Template
		Title                                                      string `json:"title"`            //nolint:revive // That's intended.
		ShortDescription                                           string `json:"shortDescription"` //nolint:revive // That's intended.
		LongDescription                                            string `json:"longDescription"`  //nolint:revive // That's intended.
		ErrorDescription                                           string `json:"errorDescription"` //nolint:revive // That's intended.
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
			Type             string  `yaml:"type" mapstructure:"type"`
			Icon             string  `yaml:"icon" mapstructure:"icon"`
			URL              string  `yaml:"url" mapstructure:"url"`
			ConfirmationCode string  `yaml:"confirmationCode" mapstructure:"confirmationCode"`
			Group            string  `yaml:"group" mapstructure:"group"`
			Prize            float64 `yaml:"prize" mapstructure:"prize"`
			RequiredQuantity uint64  `yaml:"requiredQuantity" mapstructure:"requiredQuantity"`
			GroupIndex       uint64  `yaml:"groupIndex" mapstructure:"groupIndex"`
		} `yaml:"tasksList" mapstructure:"tasksList"`
		BadgesList struct {
			Levels []*struct {
				Name string `yaml:"name" mapstructure:"name"`
			} `yaml:"levels" mapstructure:"levels"`
			Coins []*struct {
				Name string `yaml:"name" mapstructure:"name"`
			} `yaml:"coins" mapstructure:"coins"`
			Socials []*struct {
				Name string `yaml:"name" mapstructure:"name"`
			} `yaml:"socials" mapstructure:"socials"`
		} `yaml:"badgesList" mapstructure:"badgesList"`
		AdminUsers             []string                 `yaml:"adminUsers" mapstructure:"adminUsers"`
		messagebroker.Config   `mapstructure:",squash"` //nolint:tagliatelle // Nope.
		RequiredFriendsInvited uint64                   `yaml:"requiredFriendsInvited"`
		TasksV2Enabled         bool                     `yaml:"tasksV2Enabled" mapstructure:"tasksV2Enabled"`
	}
)
