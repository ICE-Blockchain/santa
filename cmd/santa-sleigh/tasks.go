// SPDX-License-Identifier: ice License 1.0

package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ice-blockchain/santa/tasks"
	"github.com/ice-blockchain/wintr/server"
)

func (s *service) setupTasksRoutes(router *server.Router) {
	router.
		Group("/v1w").
		PUT("/tasks/:taskType/users/:userId", server.RootHandler(s.PseudoCompleteTask))
}

// PseudoCompleteTask godoc
//
//	@Schemes
//	@Description	Completes the specific task (identified via task type) for the specified user.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string					true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			taskType		path	string					true	"the type of the task"		enums(claim_username,start_mining,upload_profile_picture,follow_us_on_twitter,join_twitter,join_telegram,invite_friends,invite_friends_5,invite_friends_10,invite_friends_25,invite_friends_50,invite_friends_100,invite_friends_200,join_youtube,watch_video_with_code_confirmation_1,claim_badge_l1,claim_badge_l2,claim_badge_l3,claim_badge_l4,claim_badge_l5,claim_badge_l6,claim_badge_c1,claim_badge_c2,claim_badge_c3,claim_badge_c4,claim_badge_c5,claim_badge_c6,claim_badge_c7,claim_badge_c8,claim_badge_c9,claim_badge_c10,claim_badge_s1,claim_badge_s2,claim_badge_s3,claim_badge_s4,claim_badge_s5,claim_badge_s6,claim_badge_s7,claim_badge_s8,claim_badge_s9,claim_badge_s10,claim_level_1,claim_level_2,claim_level_3,claim_level_4,claim_level_5,claim_level_6,claim_level_7,claim_level_8,claim_level_9,claim_level_10,claim_level_11,claim_level_12,claim_level_13,claim_level_14,claim_level_15,claim_level_16,claim_level_17,claim_level_18,claim_level_19,claim_level_20,claim_level_21,mining_streak_7,mining_streak_14,mining_streak_30,join_reddit_ion,join_instagram_ion,join_twitter_ion,join_telegram_ion,signup_sunwaves,signup_callfluent,signup_sauces,signup_sealsend,signup_doctorx,signup_cryptomayors,join_twitter_multiversx,join_twitter_xportal,join_telegram_multiversx,join_bullish_cmc,join_ion_cmc,join_watchlist_cmc,join_portfolio_coingecko,join_holdcoin,join_human,join_hipo,join_freedogs,join_athene,join_kolo,join_ducks,join_cmc_ton,join_cmc_sol,join_cmc_bnb,join_cmc_eth,join_cmc_btc,join_bearfi,join_boinkers,join_dejendog,join_catgoldminer,join_cmc_pnut,join_cmc_ada,join_cmc_doge,join_cmc_xrp,join_cmc_act,join_tonkombat,join_tonai,join_pigs,join_capybara,join_sidekick,join_iceberg,join_goats,join_tapcoins,join_tokyobeast,join_twitter_pichain,join_sugar)
//	@Param			userId			path	string					true	"the id of the user that completed the task"
//	@Param			request			body	CompleteTaskRequestBody	false	"Request params. Set it only if task completion requires additional data."
//	@Success		200				"ok"
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		403				{object}	server.ErrorResponse	"if not allowed"
//	@Failure		404				{object}	server.ErrorResponse	"if user not found"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1w/tasks/{taskType}/users/{userId} [PUT].
func (s *service) PseudoCompleteTask( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[CompleteTaskRequestBody, any],
) (*server.Response[any], *server.Response[server.ErrorResponse]) {
	task := &tasks.Task{
		Data:   req.Data.Data,
		Type:   req.Data.TaskType,
		UserID: req.AuthenticatedUser.UserID,
	}
	if err := s.tasksProcessor.PseudoCompleteTask(ctx, task, req.Data.DryRun); err != nil {
		err = errors.Wrapf(err, "failed to PseudoCompleteTask for %#v, userID:%v", req.Data, req.AuthenticatedUser.UserID)
		switch {
		case errors.Is(err, tasks.ErrInvalidSocialProperties):
			return nil, server.UnprocessableEntity(err, invalidPropertiesErrorCode)
		case errors.Is(err, tasks.ErrRelationNotFound):
			return nil, server.NotFound(err, userNotFoundErrorCode)
		case errors.Is(err, tasks.ErrTaskNotCompleted):
			return nil, server.NotFound(err, taskNotCompletedCode)
		default:
			return nil, server.Unexpected(err)
		}
	}

	return server.OK[any](), nil
}
