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
		UserID   string `uri:"userId" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4" swaggerignore:"true" required:"true"`
		Language string `uri:"language" example:"en" swaggerignore:"true" required:"false"`
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
		GET("/achievement-summaries/badges/users/:userId", server.RootHandler(s.GetBadgeSummary))
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
	resp, err := s.tasksProcessor.GetTasks(ctx, req.Data.UserID, req.Data.Language)
	if err != nil {
		err = errors.Wrapf(err, "failed to GetTasks for data:%#v", req.Data)

		return nil, server.Unexpected(err)
	}

	return server.OK(&resp), nil
}
