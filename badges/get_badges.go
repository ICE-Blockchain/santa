// SPDX-License-Identifier: ice License 1.0

package badges

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
)

func (r *repository) GetBadges(ctx context.Context, groupType GroupType, userID string) ([]*Badge, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	stats, err := r.getBadgesDistributon(ctx, groupType)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to getStatistics for %v", groupType)
	}
	userProgress, err := r.getProgress(ctx, userID, true)
	if err != nil && !errors.Is(err, ErrRelationNotFound) {
		return nil, errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	if userProgress != nil && (userProgress.HideBadges && requestingUserID(ctx) != userID) {
		return nil, ErrHidden
	}

	return userProgress.buildBadges(groupType, stats), nil
}

func (r *repository) GetSummary(ctx context.Context, userID string) ([]*BadgeSummary, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	userProgress, err := r.getProgress(ctx, userID, true)
	if err != nil && !errors.Is(err, ErrRelationNotFound) {
		return nil, errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	if userProgress != nil && (userProgress.HideBadges && requestingUserID(ctx) != userID) {
		return nil, ErrHidden
	}

	return userProgress.buildBadgeSummaries(), nil
}

//nolint:revive // .
func (r *repository) getProgress(ctx context.Context, userID string, tolerateOldData bool) (res *progress, err error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	sql := `SELECT * FROM badge_progress WHERE user_id = $1`
	if tolerateOldData {
		res, err = storage.Get[progress](ctx, r.db, sql, userID)
	} else {
		res, err = storage.ExecOne[progress](ctx, r.db, sql, userID)
	}

	if res == nil {
		return nil, ErrRelationNotFound
	}

	return res, errors.Wrapf(err, "can't get badge progress for userID:%v", userID)
}

func (r *repository) getBadgesDistributon(ctx context.Context, groupType GroupType) (map[Type]float64, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	sql := fmt.Sprintf(`SELECT COALESCE(COUNT(user_id), 0) AS count,
				(CASE
					%v
				END) AS range
				FROM badge_progress
				GROUP BY range
				ORDER BY range ASC`, strings.Join(r.preparePercentageDistributionSQL(groupType), "\n"))
	resp, err := storage.Select[badgeDistribution](ctx, r.db, sql)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BADGE_STATISTICS for groupType:%v", groupType)
	}

	return r.calculatePercentages(resp, groupType), nil
}

//nolint:funlen,gocognit,revive // .
func (r *repository) preparePercentageDistributionSQL(groupType GroupType) []string {
	var whenValues []string
	switch groupType {
	case CoinGroupType:
		whenValues = make([]string, 0, len(r.cfg.Coins))
		for idx, rnge := range r.cfg.Coins {
			val := fmt.Sprintf("when balance >= %v", rnge.FromInclusive)
			if rnge.ToInclusive != 0 {
				val = fmt.Sprintf("%v AND balance <= %v", val, rnge.ToInclusive)
			}
			whenValues = append(whenValues, fmt.Sprintf("%v THEN 'c%v'", val, idx+1))
		}
	case LevelGroupType:
		whenValues = make([]string, 0, len(r.cfg.Levels))
		for idx, rnge := range r.cfg.Levels {
			val := fmt.Sprintf("when completed_levels >= %v", rnge.FromInclusive)
			if rnge.ToInclusive != 0 {
				val = fmt.Sprintf("%v AND completed_levels <= %v", val, rnge.ToInclusive)
			}
			whenValues = append(whenValues, fmt.Sprintf("%v THEN 'l%v'", val, idx+1))
		}
	case SocialGroupType:
		whenValues = make([]string, 0, len(r.cfg.Socials))
		for idx, rnge := range r.cfg.Socials {
			val := fmt.Sprintf("when friends_invited >= %v", rnge.FromInclusive)
			if rnge.ToInclusive != 0 {
				val = fmt.Sprintf("%v AND friends_invited <= %v", val, rnge.ToInclusive)
			}
			whenValues = append(whenValues, fmt.Sprintf("%v THEN 's%v'", val, idx+1))
		}
	}

	return whenValues
}

func (r *repository) calculatePercentages(distribution []*badgeDistribution, groupType GroupType) map[Type]float64 {
	var length int
	switch groupType {
	case CoinGroupType:
		length = len(r.cfg.Coins)
	case LevelGroupType:
		length = len(r.cfg.Levels)
	case SocialGroupType:
		length = len(r.cfg.Socials)
	}
	var totalUsers uint64
	for _, val := range distribution {
		totalUsers += val.Count
	}
	result := make(map[Type]float64, length)
	for _, val := range distribution {
		result[Type(val.Range)] = float64(val.Count) / float64(totalUsers) * 100
	}

	return result
}

func (p *progress) buildBadges(groupType GroupType, stats map[Type]float64) []*Badge {
	resp := make([]*Badge, 0, len(AllGroups[groupType]))
	for _, badgeType := range AllGroups[groupType] {
		resp = append(resp, &Badge{
			AchievingRange:              Milestones[badgeType],
			Name:                        AllNames[groupType][badgeType],
			Type:                        badgeType,
			GroupType:                   groupType,
			PercentageOfUsersInProgress: stats[badgeType],
		})
	}
	if p == nil || p.AchievedBadges == nil || len(*p.AchievedBadges) == 0 {
		return resp
	}
	achievedBadges := make(map[Type]bool, len(resp))
	for _, achievedBadge := range *p.AchievedBadges {
		achievedBadges[achievedBadge] = true
	}
	for _, badge := range resp {
		badge.Achieved = achievedBadges[badge.Type]
	}

	return resp
}

func (p *progress) buildBadgeSummaries() []*BadgeSummary { //nolint:gocognit,revive // .
	resp := make([]*BadgeSummary, 0, len(AllGroups))
	for _, groupType := range &GroupsOrderSummaries {
		types := AllGroups[groupType]
		lastAchievedIndex := 0
		if p != nil && p.AchievedBadges != nil {
			for ix, badgeType := range types {
				for _, achievedBadge := range *p.AchievedBadges {
					if badgeType == achievedBadge {
						lastAchievedIndex = ix
					}
				}
			}
		}
		resp = append(resp, &BadgeSummary{
			Name:      AllNames[groupType][types[lastAchievedIndex]],
			GroupType: groupType,
			Index:     uint64(lastAchievedIndex),
			LastIndex: uint64(len(types) - 1),
		})
	}

	return resp
}
