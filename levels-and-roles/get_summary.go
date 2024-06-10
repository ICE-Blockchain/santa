// SPDX-License-Identifier: ice License 1.0

package levelsandroles

import (
	"context"

	"github.com/pkg/errors"

	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
)

func (r *repository) GetSummary(ctx context.Context, userID string) (*Summary, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	if res, err := r.getProgress(ctx, userID, true); err != nil && !errors.Is(err, storage.ErrRelationNotFound) {
		return nil, errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	} else { //nolint:revive // .
		return r.newSummary(res, requestingUserID(ctx)), nil
	}
}

//nolint:revive // .
func (r *repository) getProgress(ctx context.Context, userID string, tolerateOldData bool) (res *progress, err error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	sql := `SELECT *
			FROM levels_and_roles_progress
			WHERE user_id = $1`
	if tolerateOldData {
		res, err = storage.Get[progress](ctx, r.db, sql, userID)
	} else {
		res, err = storage.ExecOne[progress](ctx, r.db, sql, userID)
	}

	if errors.Is(err, storage.ErrNotFound) {
		return nil, storage.ErrRelationNotFound
	}

	return
}

func (r *repository) newSummary(pr *progress, requestingUserID string) *Summary {
	var level uint64
	if pr == nil || !pr.HideLevel || requestingUserID == pr.UserID {
		level = pr.level()
	}
	var roles []*Role
	if pr == nil || !pr.HideRole || requestingUserID == pr.UserID {
		roles = pr.roles(r)
	}

	return &Summary{Roles: roles, Level: level}
}

func (p *progress) level() uint64 {
	if p == nil || p.CompletedLevels == nil || len(*p.CompletedLevels) == 0 {
		return 1
	} else { //nolint:revive // .
		return uint64(len(*p.CompletedLevels))
	}
}

func (p *progress) roles(repo *repository) []*Role {
	if p == nil || p.EnabledRoles == nil || len(*p.EnabledRoles) == 0 {
		return []*Role{
			{
				Type:    repo.cfg.RoleNames[SnowmanRoleIndex],
				Enabled: true,
			},
			{
				Type:    repo.cfg.RoleNames[AmbassadorRoleIndex],
				Enabled: false,
			},
		}
	} else { //nolint:revive // .
		return []*Role{
			{
				Type:    repo.cfg.RoleNames[SnowmanRoleIndex],
				Enabled: false,
			},
			{
				Type:    repo.cfg.RoleNames[AmbassadorRoleIndex],
				Enabled: true,
			},
		}
	}
}
