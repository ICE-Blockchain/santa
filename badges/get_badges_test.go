// SPDX-License-Identifier: ice License 1.0

package badges

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ice-blockchain/eskimo/users"
)

//nolint:funlen,paralleltest,tparallel // A lot of testcases. Not needed to be the parallel due to loads the global variables.
func Test_Progress_BuildBadges(t *testing.T) {
	repo := &repository{cfg: defaultCfg()}
	loadBadges(repo.cfg)
	tests := []*struct {
		progress *progress
		stats    map[Type]float64
		group    GroupType
		name     string
		expected []*Badge
	}{
		{
			name:     "Nothing achieved",
			progress: badgeProgress(nil, 0, 0, 0),
			group:    LevelGroupType,
			stats:    map[Type]float64{},
			expected: []*Badge{
				expectedBadge(Level1Type, false, 0),
				expectedBadge(Level2Type, false, 0),
				expectedBadge(Level3Type, false, 0),
				expectedBadge(Level4Type, false, 0),
				expectedBadge(Level5Type, false, 0),
				expectedBadge(Level6Type, false, 0),
			},
		},
		{
			name: "partially achieved",
			progress: badgeProgress(&users.Enum[Type]{
				Level1Type, Level2Type, Level3Type,
			}, 0, 0, 0),
			group: LevelGroupType,
			stats: map[Type]float64{
				Level1Type: 50.0,
				Level2Type: 30.0,
				Level3Type: 20.0,
			},
			expected: []*Badge{
				expectedBadge(Level1Type, true, 50.0),
				expectedBadge(Level2Type, true, 30.0),
				expectedBadge(Level3Type, true, 20.0),
				expectedBadge(Level4Type, false, 0),
				expectedBadge(Level5Type, false, 0),
				expectedBadge(Level6Type, false, 0),
			},
		},
		{
			name: "all achieved, preserve order",
			progress: badgeProgress(&users.Enum[Type]{
				Coin1Type, Coin2Type, Coin3Type, Coin5Type, Coin4Type,
				Coin10Type, Coin6Type, Coin7Type, Coin8Type, Coin9Type,
			}, 0, 0, 0),
			group: CoinGroupType,
			stats: map[Type]float64{
				Coin1Type:  10,
				Coin2Type:  10,
				Coin3Type:  10,
				Coin4Type:  10,
				Coin5Type:  10,
				Coin6Type:  10,
				Coin7Type:  10,
				Coin8Type:  10,
				Coin9Type:  10,
				Coin10Type: 10,
			},
			expected: []*Badge{
				expectedBadge(Coin1Type, true, 10),
				expectedBadge(Coin2Type, true, 10),
				expectedBadge(Coin3Type, true, 10),
				expectedBadge(Coin4Type, true, 10),
				expectedBadge(Coin5Type, true, 10),
				expectedBadge(Coin6Type, true, 10),
				expectedBadge(Coin7Type, true, 10),
				expectedBadge(Coin8Type, true, 10),
				expectedBadge(Coin9Type, true, 10),
				expectedBadge(Coin10Type, true, 10),
			},
		},
		{
			name:     "empty statistics",
			progress: badgeProgress(&users.Enum[Type]{Coin1Type}, 0, 0, 0),
			group:    CoinGroupType,
			stats:    map[Type]float64{},
			expected: []*Badge{
				expectedBadge(Coin1Type, true, 0),
				expectedBadge(Coin2Type, false, 0),
				expectedBadge(Coin3Type, false, 0),
				expectedBadge(Coin4Type, false, 0),
				expectedBadge(Coin5Type, false, 0),
				expectedBadge(Coin6Type, false, 0),
				expectedBadge(Coin7Type, false, 0),
				expectedBadge(Coin8Type, false, 0),
				expectedBadge(Coin9Type, false, 0),
				expectedBadge(Coin10Type, false, 0),
			},
		},
		{
			name:     "some are achieved but in another group",
			progress: badgeProgress(&users.Enum[Type]{Coin1Type, Level1Type}, 0, 0, 0),
			group:    SocialGroupType,
			stats: map[Type]float64{
				Social1Type: 50,
				Social2Type: 50,
			},
			expected: []*Badge{
				expectedBadge(Social1Type, false, 50),
				expectedBadge(Social2Type, false, 50),
				expectedBadge(Social3Type, false, 0),
				expectedBadge(Social4Type, false, 0),
				expectedBadge(Social5Type, false, 0),
				expectedBadge(Social6Type, false, 0),
				expectedBadge(Social7Type, false, 0),
				expectedBadge(Social8Type, false, 0),
				expectedBadge(Social9Type, false, 0),
				expectedBadge(Social10Type, false, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := tt.progress.buildBadges(tt.group, tt.stats)
			assert.Equal(t, tt.expected, actual, tt.name)
		})
	}
}

//nolint:funlen,paralleltest,tparallel // A lot of testcases. Not needed to be parallel due to load the global variables.
func Test_Progress_BuildSummary(t *testing.T) {
	loadBadges(defaultCfg())
	tests := []*struct {
		name              string
		progress          *progress
		expectedSummaries []*BadgeSummary
	}{
		{
			"Nothning achieved",
			badgeProgress(nil, 0, 0, 0),
			[]*BadgeSummary{
				expectedBadgeSummaryFromBadgeType(Level1Type),
				expectedBadgeSummaryFromBadgeType(Coin1Type),
				expectedBadgeSummaryFromBadgeType(Social1Type),
			},
		},
		{
			"Automatically achieved",
			badgeProgress(&users.Enum[Type]{
				Level1Type,
				Coin1Type,
				Social1Type,
			}, 0, 0, 0),
			[]*BadgeSummary{
				expectedBadgeSummaryFromBadgeType(Level1Type),
				expectedBadgeSummaryFromBadgeType(Coin1Type),
				expectedBadgeSummaryFromBadgeType(Social1Type),
			},
		},
		{
			"In some groups achieved more than one",
			badgeProgress(&users.Enum[Type]{
				Level1Type,
				Level2Type,
				Level3Type,
				Coin1Type,
				Coin2Type,
				Coin3Type,
				Coin4Type,
				Coin5Type,
				Social1Type,
			}, 0, 0, 0),
			[]*BadgeSummary{
				expectedBadgeSummaryFromBadgeType(Level3Type),
				expectedBadgeSummaryFromBadgeType(Coin5Type),
				expectedBadgeSummaryFromBadgeType(Social1Type),
			},
		},
		{
			"One group is complete",
			badgeProgress(&users.Enum[Type]{
				Level1Type,
				Level2Type,
				Level3Type,
				Level4Type,
				Level5Type,
				Level6Type,
				Coin1Type,
				Social1Type,
			}, 0, 0, 0),
			[]*BadgeSummary{
				expectedBadgeSummaryFromBadgeType(Level6Type),
				expectedBadgeSummaryFromBadgeType(Coin1Type),
				expectedBadgeSummaryFromBadgeType(Social1Type),
			},
		},
		{
			"All groups are completed",
			badgeProgress(&users.Enum[Type]{
				Level1Type, Level2Type, Level3Type, Level4Type, Level5Type, Level6Type,
				Coin1Type, Coin2Type, Coin3Type, Coin4Type, Coin5Type, Coin6Type, Coin7Type, Coin8Type, Coin9Type, Coin10Type,
				Social1Type, Social2Type, Social3Type, Social4Type, Social5Type,
				Social6Type, Social7Type, Social8Type, Social9Type, Social10Type,
			}, 0, 0, 0),
			[]*BadgeSummary{
				expectedBadgeSummaryFromBadgeType(Level6Type),
				expectedBadgeSummaryFromBadgeType(Coin10Type),
				expectedBadgeSummaryFromBadgeType(Social10Type),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actualSummaries := tt.progress.buildBadgeSummaries()
			assert.ElementsMatch(t, tt.expectedSummaries, actualSummaries, tt.name)
		})
	}
}

func expectedBadge(badgeType Type, achieved bool, percent float64) *Badge {
	return &Badge{
		Name:                        AllNames[GroupTypeForEachType[badgeType]][badgeType],
		Type:                        badgeType,
		GroupType:                   GroupTypeForEachType[badgeType],
		PercentageOfUsersInProgress: percent,
		Achieved:                    achieved,
		AchievingRange: AchievingRange{
			Name:          AllNames[GroupTypeForEachType[badgeType]][badgeType],
			FromInclusive: Milestones[badgeType].FromInclusive,
			ToInclusive:   Milestones[badgeType].ToInclusive,
		},
	}
}

func expectedBadgeSummary(name string, group GroupType, index, lastIndex uint64) *BadgeSummary {
	return &BadgeSummary{
		Name:      name,
		GroupType: group,
		Index:     index,
		LastIndex: lastIndex,
	}
}

func expectedBadgeSummaryFromBadgeType(lastAchievedType Type) *BadgeSummary {
	group := GroupTypeForEachType[lastAchievedType]
	var order int
	switch group {
	case SocialGroupType:
		order = SocialTypeOrder[lastAchievedType]
	case LevelGroupType:
		order = LevelTypeOrder[lastAchievedType]
	case CoinGroupType:
		order = CoinTypeOrder[lastAchievedType]
	}

	return expectedBadgeSummary(AllNames[group][lastAchievedType], group, uint64(order), uint64(len(AllGroups[group])-1))
}

//nolint:funlen // .
func TestPreparePercentageDistributionSQL(t *testing.T) {
	t.Parallel()
	repo := &repository{cfg: defaultCfg()}
	loadBadges(repo.cfg)
	vals := repo.preparePercentageDistributionSQL(LevelGroupType)
	assert.EqualValues(t, []string{
		"when completed_levels >= 0 AND completed_levels <= 1 THEN 'l1'",
		"when completed_levels >= 2 AND completed_levels <= 3 THEN 'l2'",
		"when completed_levels >= 4 AND completed_levels <= 5 THEN 'l3'",
		"when completed_levels >= 6 AND completed_levels <= 7 THEN 'l4'",
		"when completed_levels >= 8 AND completed_levels <= 9 THEN 'l5'",
		"when completed_levels >= 10 THEN 'l6'",
	}, vals)
	vals = repo.preparePercentageDistributionSQL(SocialGroupType)
	assert.EqualValues(t, []string{
		"when friends_invited >= 0 AND friends_invited <= 1 THEN 's1'",
		"when friends_invited >= 2 AND friends_invited <= 3 THEN 's2'",
		"when friends_invited >= 4 AND friends_invited <= 5 THEN 's3'",
		"when friends_invited >= 6 AND friends_invited <= 7 THEN 's4'",
		"when friends_invited >= 8 AND friends_invited <= 9 THEN 's5'",
		"when friends_invited >= 10 AND friends_invited <= 11 THEN 's6'",
		"when friends_invited >= 12 AND friends_invited <= 13 THEN 's7'",
		"when friends_invited >= 14 AND friends_invited <= 15 THEN 's8'",
		"when friends_invited >= 16 AND friends_invited <= 17 THEN 's9'",
		"when friends_invited >= 18 THEN 's10'",
	}, vals)
	vals = repo.preparePercentageDistributionSQL(CoinGroupType)
	assert.EqualValues(t, []string{
		"when balance >= 0 AND balance <= 10 THEN 'c1'",
		"when balance >= 20 AND balance <= 30 THEN 'c2'",
		"when balance >= 40 AND balance <= 50 THEN 'c3'",
		"when balance >= 60 AND balance <= 70 THEN 'c4'",
		"when balance >= 80 AND balance <= 90 THEN 'c5'",
		"when balance >= 100 AND balance <= 110 THEN 'c6'",
		"when balance >= 120 AND balance <= 130 THEN 'c7'",
		"when balance >= 140 AND balance <= 150 THEN 'c8'",
		"when balance >= 160 AND balance <= 170 THEN 'c9'",
		"when balance >= 180 THEN 'c10'",
	}, vals)
}
