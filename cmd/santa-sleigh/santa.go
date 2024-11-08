// SPDX-License-Identifier: ice License 1.0

package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ice-blockchain/santa/badges"
	levelsandroles "github.com/ice-blockchain/santa/levels-and-roles"
	"github.com/ice-blockchain/santa/tasks"
	"github.com/ice-blockchain/wintr/server"
)

// Public API.

type (
	GetTasksArg struct {
		UserID   string           `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" swaggerignore:"true" required:"true"`
		Language string           `form:"language" example:"en" swaggerignore:"true" required:"false"`
		Status   tasks.TaskStatus `form:"status" example:"pending" swaggerignore:"true" required:"false" enums:"pending,completed"`
	}
	GetTaskArg struct {
		UserID   string     `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" swaggerignore:"true" required:"true"`
		Language string     `form:"language" example:"en" swaggerignore:"true" required:"false"`
		TaskType tasks.Type `uri:"taskType" example:"claim_username" swaggerignore:"true" required:"true" enums:"level,coin,social"`
	}
	GetLevelsAndRolesSummaryArg struct {
		UserID string `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" allowForbiddenGet:"true" swaggerignore:"true" required:"true"`
	}
	GetBadgeSummaryArg struct {
		UserID string `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" allowForbiddenGet:"true" swaggerignore:"true" required:"true"`
	}
	GetBadgesArg struct {
		UserID    string           `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" allowForbiddenGet:"true" swaggerignore:"true" required:"true"`
		GroupType badges.GroupType `uri:"badgeType" example:"social" swaggerignore:"true" required:"true" enums:"level,coin,social"`
	}
)

// Private API.

// Values for server.ErrorResponse#Code.
const (
	badgesHiddenErrorCode = "BADGES_HIDDEN"
)

func (s *service) registerReadRoutes(router *server.Router) {
	s.setupTasksReadRoutes(router)
	s.setupLevelsAndRolesReadRoutes(router)
	s.setupBadgesReadRoutes(router)
}

func (s *service) setupBadgesReadRoutes(router *server.Router) {
	router.
		Group("/v1r").
		GET("/badges/:badgeType/users/:userId", server.RootHandler(s.GetBadges)).
		GET("/achievement-summaries/badges/users/:userId", server.RootHandler(s.GetBadgeSummary)).
		GET("/tasks/:taskType/users/:userId", server.RootHandler(s.GetTask))
}

// GetBadges godoc
//
//	@Schemes
//	@Description	Returns all badges of the specific type for the user, with the progress for each of them.
//	@Tags			Badges
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			userId			path		string	true	"the id of the user you need progress for"
//	@Param			badgeType		path		string	true	"the type of the badges"	enums(level,coin,social)
//	@Success		200				{array}		badges.Badge
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		403				{object}	server.ErrorResponse	"if not allowed"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1r/badges/{badgeType}/users/{userId} [GET].
func (s *service) GetBadges( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[GetBadgesArg, []*badges.Badge],
) (*server.Response[[]*badges.Badge], *server.Response[server.ErrorResponse]) {
	resp, err := s.badgesProcessor.GetBadges(ctx, req.Data.GroupType, req.Data.UserID)
	if err != nil {
		err = errors.Wrapf(err, "failed to GetBadges for data:%#v", req.Data)
		if errors.Is(err, badges.ErrHidden) {
			return nil, server.ForbiddenWithCode(err, badgesHiddenErrorCode)
		}

		return nil, server.Unexpected(err)
	}

	return server.OK(&resp), nil
}

// GetBadgeSummary godoc
//
//	@Schemes
//	@Description	Returns user's summary about badges.
//	@Tags			Badges
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			userId			path		string	true	"the id of the user you need summary for"
//	@Success		200				{array}		badges.BadgeSummary
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		403				{object}	server.ErrorResponse	"if not allowed"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1r/achievement-summaries/badges/users/{userId} [GET].
func (s *service) GetBadgeSummary( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[GetBadgeSummaryArg, []*badges.BadgeSummary],
) (*server.Response[[]*badges.BadgeSummary], *server.Response[server.ErrorResponse]) {
	resp, err := s.badgesProcessor.GetSummary(ctx, req.Data.UserID)
	if err != nil {
		err = errors.Wrapf(err, "failed to badges.GetSummary for data:%#v", req.Data)
		if errors.Is(err, badges.ErrHidden) {
			return nil, server.ForbiddenWithCode(err, badgesHiddenErrorCode)
		}

		return nil, server.Unexpected(err)
	}

	return server.OK(&resp), nil
}

func (s *service) setupLevelsAndRolesReadRoutes(router *server.Router) {
	router.
		Group("/v1r").
		GET("/achievement-summaries/levels-and-roles/users/:userId", server.RootHandler(s.GetLevelsAndRolesSummary))
}

// GetLevelsAndRolesSummary godoc
//
//	@Schemes
//	@Description	Returns user's summary about levels & roles.
//	@Tags			Levels & Roles
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			userId			path		string	true	"the id of the user you need summary for"
//	@Success		200				{object}	levelsandroles.Summary
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1r/achievement-summaries/levels-and-roles/users/{userId} [GET].
func (s *service) GetLevelsAndRolesSummary( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[GetLevelsAndRolesSummaryArg, levelsandroles.Summary],
) (*server.Response[levelsandroles.Summary], *server.Response[server.ErrorResponse]) {
	resp, err := s.levelsAndRolesProcessor.GetSummary(ctx, req.Data.UserID)
	if err != nil {
		err = errors.Wrapf(err, "failed to levelsandroles.GetSummary for data:%#v", req.Data)

		return nil, server.Unexpected(err)
	}

	return server.OK(resp), nil
}

func (s *service) setupTasksReadRoutes(router *server.Router) {
	router.
		Group("/v1r").
		GET("/tasks/x/users/:userId", server.RootHandler(s.GetTasks))
}

// GetTasks godoc
//
//	@Schemes
//	@Description	Returns all the tasks and provided user's progress for each of them.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			userId			path		string	true	"the id of the user you need progress for"
//	@Param			language		query		string	false	"language of translations for tasks description"
//	@Param			status			query		string	false	"pending/completed status filter"	enums(pending,completed)
//	@Success		200				{array}		tasks.Task
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		403				{object}	server.ErrorResponse	"if not allowed"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1r/tasks/x/users/{userId} [GET].
func (s *service) GetTasks( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[GetTasksArg, []*tasks.Task],
) (*server.Response[[]*tasks.Task], *server.Response[server.ErrorResponse]) {
	if req.Data.UserID != req.AuthenticatedUser.UserID {
		return nil, server.Forbidden(errors.Errorf("not allowed. %v != %v", req.Data.UserID, req.AuthenticatedUser.UserID))
	}
	resp, err := s.tasksProcessor.GetTasks(ctx, req.Data.UserID, req.Data.Language, req.Data.Status)
	if err != nil {
		err = errors.Wrapf(err, "failed to GetTasks for data:%#v", req.Data)
		switch {
		case errors.Is(err, tasks.ErrWrongRequestedTasksStatus):
			return nil, server.BadRequest(err, invalidPropertiesErrorCode)
		default:
			return nil, server.Unexpected(err)
		}
	}

	return server.OK(&resp), nil
}

// GetTask godoc
//
//	@Schemes
//	@Description	Returns the tasks and provided user's progress for specific task.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Insert your access token"	default(Bearer <Add access token here>)
//	@Param			taskType		path		string	true	"the type of the task"		enums(claim_username,start_mining,upload_profile_picture,join_twitter,join_telegram,invite_friends_5,invite_friends_10,invite_friends_25,invite_friends_50,invite_friends_100,invite_friends_200,join_youtube,watch_video_with_code_confirmation_1,claim_badge_l1,claim_badge_l2,claim_badge_l3,claim_badge_l4,claim_badge_l5,claim_badge_l6,claim_badge_c1,claim_badge_c2,claim_badge_c3,claim_badge_c4,claim_badge_c5,claim_badge_c6,claim_badge_c7,claim_badge_c8,claim_badge_c9,claim_badge_c10,claim_badge_s1,claim_badge_s2,claim_badge_s3,claim_badge_s4,claim_badge_s5,claim_badge_s6,claim_badge_s7,claim_badge_s8,claim_badge_s9,claim_badge_s10,claim_level_1,claim_level_2,claim_level_3,claim_level_4,claim_level_5,claim_level_6,claim_level_7,claim_level_8,claim_level_9,claim_level_10,claim_level_11,claim_level_12,claim_level_13,claim_level_14,claim_level_15,claim_level_16,claim_level_17,claim_level_18,claim_level_19,claim_level_20,claim_level_21,mining_streak_7,mining_streak_14,mining_streak_30,join_reddit_ion,join_instagram_ion,join_twitter_ion,join_telegram_ion,signup_sunwaves,signup_callfluent,signup_sauces,signup_sealsend,singup_doctorx,signup_tokero,join_twitter_multiversx,join_twitter_xportal,join_telegram_multiversx,join_bullish_cmc,join_ion_cmc,join_watchlist_cmc,join_portfolio_coingecko,join_holdcoin,join_human,join_hipo,join_freedogs,join_athene,join_kolo,join_ducks)
//	@Param			userId			path		string	true	"the id of the user you need progress for"
//	@Param			language		query		string	false	"language of translations for task description"
//	@Success		200				{array}		tasks.Task
//	@Failure		400				{object}	server.ErrorResponse	"if validations fail"
//	@Failure		401				{object}	server.ErrorResponse	"if not authorized"
//	@Failure		403				{object}	server.ErrorResponse	"if not allowed"
//	@Failure		422				{object}	server.ErrorResponse	"if syntax fails"
//	@Failure		500				{object}	server.ErrorResponse
//	@Failure		504				{object}	server.ErrorResponse	"if request times out"
//	@Router			/v1r/tasks/{taskType}/users/{userId} [GET].
func (s *service) GetTask( //nolint:gocritic // False negative.
	ctx context.Context,
	req *server.Request[GetTaskArg, *tasks.Task],
) (*server.Response[*tasks.Task], *server.Response[server.ErrorResponse]) {
	if req.Data.UserID != req.AuthenticatedUser.UserID {
		return nil, server.Forbidden(errors.Errorf("not allowed. %v != %v", req.Data.UserID, req.AuthenticatedUser.UserID))
	}
	resp, err := s.tasksProcessor.GetTask(ctx, req.Data.UserID, req.Data.Language, req.Data.TaskType)
	if err != nil {
		switch {
		case errors.Is(err, tasks.ErrNotFound):
			return nil, server.NotFound(err, taskNotFoundCode)
		case errors.Is(err, tasks.ErrNotSupported):
			return nil, server.Forbidden(err)
		default:
			return nil, server.Unexpected(errors.Wrapf(err, "failed to GetTask for data:%#v", req.Data))
		}
	}

	return server.OK(&resp), nil
}
