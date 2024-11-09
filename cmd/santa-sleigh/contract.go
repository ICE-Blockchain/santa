// SPDX-License-Identifier: ice License 1.0

package main

import (
	"github.com/ice-blockchain/santa/badges"
	friendsinvited "github.com/ice-blockchain/santa/friends-invited"
	levelsandroles "github.com/ice-blockchain/santa/levels-and-roles"
	"github.com/ice-blockchain/santa/tasks"
)

// Public API.

type (
	CompleteTaskRequestBody struct {
		Data     *tasks.Data `json:"data,omitempty"`
		UserID   string      `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" swaggerignore:"true" required:"true"`
		TaskType tasks.Type  `uri:"taskType" example:"start_mining" swaggerignore:"true" required:"true" enums:"claim_username,start_mining,upload_profile_picture,follow_us_on_twitter,join_twitter,join_telegram,invite_friends,invite_friends_5,invite_friends_10,invite_friends_25,invite_friends_50,invite_friends_100,invite_friends_200,join_youtube,watch_video_with_code_confirmation_1,invite_friends_5,invite_friends_10,claim_badge_l1,claim_badge_l2,claim_badge_l3,claim_badge_l4,claim_badge_l5,claim_badge_l6,claim_badge_c1,claim_badge_c2,claim_badge_c3,claim_badge_c4,claim_badge_c5,claim_badge_c6,claim_badge_c7,claim_badge_c8,claim_badge_c9,claim_badge_c10,claim_badge_s1,claim_badge_s2,claim_badge_s3,claim_badge_s4,claim_badge_s5,claim_badge_s6,claim_badge_s7,claim_badge_s8,claim_badge_s9,claim_badge_s10,claim_level_1,claim_level_2,claim_level_3,claim_level_4,claim_level_5,claim_level_6,claim_level_7,claim_level_8,claim_level_9,claim_level_10,claim_level_11,claim_level_12,claim_level_13,claim_level_14,claim_level_15,claim_level_16,claim_level_17,claim_level_18,claim_level_19,claim_level_20,claim_level_21,mining_streak_7,mining_streak_14,mining_streak_30,join_reddit_ion,join_instagram_ion,join_twitter_ion,join_telegram_ion,signup_sunwaves,signup_callfluent,signup_sauces,signup_sealsend,singup_doctorx,signup_tokero,join_twitter_multiversx,join_twitter_xportal,join_telegram_multiversx,join_bullish_cmc,join_ion_cmc,join_watchlist_cmc,join_portfolio_coingecko,join_holdcoin,join_human,join_hipo,join_freedogs,join_athene,join_kolo,join_ducks,join_cmc_ton,join_cmc_sol,join_cmc_bnb,join_cmc_eth,join_cmc_btc,join_bearfi"` //nolint:lll // .
		DryRun   bool        `form:"dryRun" example:"true" swaggerignore:"true" required:"false"`
	}
)

// Private API.

const (
	applicationYamlKey = "cmd/santa-sleigh"
	swaggerRootSuffix  = "/achievements/w"
)

// Values for server.ErrorResponse#Code.
const (
	userNotFoundErrorCode      = "USER_NOT_FOUND"
	invalidPropertiesErrorCode = "INVALID_PROPERTIES"
	taskNotFoundCode           = "TASK_NOT_FOUND"
	taskNotCompletedCode       = "TASK_NOT_COMPLETED"
)

type (
	// | service implements server.State and is responsible for managing the state and lifecycle of the package.
	service struct {
		tasksProcessor          tasks.Processor
		levelsAndRolesProcessor levelsandroles.Processor
		badgesProcessor         badges.Processor
		friendsProcessor        friendsinvited.Processor
	}
	config struct {
		Host    string `yaml:"host"`
		Version string `yaml:"version"`
		Tenant  string `yaml:"tenant"`
	}
)
