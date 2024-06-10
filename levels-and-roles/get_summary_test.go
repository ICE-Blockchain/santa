// SPDX-License-Identifier: ice License 1.0

package levelsandroles

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ice-blockchain/eskimo/users"
	appcfg "github.com/ice-blockchain/wintr/config"
)

func TestNewSummary(t *testing.T) { //nolint:funlen // It's a test function
	t.Parallel()
	cfg := defaultCfg()
	repo := repository{cfg: cfg}
	testCases := []struct {
		pr               *progress
		expected         *Summary
		requestingUserID string
		name             string
	}{
		{
			name: "return 1 level and snowman role, when there is no progress",
			expected: &Summary{Level: 1, Roles: []*Role{
				{Type: cfg.RoleNames[SnowmanRoleIndex], Enabled: true},
				{Type: cfg.RoleNames[AmbassadorRoleIndex], Enabled: false},
			}},
		},
		{
			name: "return 1 level and snowman role, when roles aren't hidden",
			pr:   &progress{HideLevel: false, HideRole: false},
			expected: &Summary{Level: 1, Roles: []*Role{
				{Type: cfg.RoleNames[SnowmanRoleIndex], Enabled: true},
				{Type: cfg.RoleNames[AmbassadorRoleIndex], Enabled: false},
			}},
		},
		{
			name:             "return 1 level and snowman role, when requesting user id is same as progress user id",
			pr:               &progress{UserID: "bogus"},
			requestingUserID: "bogus",
			expected: &Summary{Level: 1, Roles: []*Role{
				{Type: cfg.RoleNames[SnowmanRoleIndex], Enabled: true},
				{Type: cfg.RoleNames[AmbassadorRoleIndex], Enabled: false},
			}},
		},
		{
			name: "return 1 level and ambassador role, when roles aren't enabled",
			pr:   &progress{EnabledRoles: &users.Enum[RoleType]{AmbassadorRoleType}},
			expected: &Summary{Level: 1, Roles: []*Role{
				{Type: cfg.RoleNames[SnowmanRoleIndex], Enabled: false},
				{Type: cfg.RoleNames[AmbassadorRoleIndex], Enabled: true},
			}},
		},
		{
			name: "return 2 level and snowman role, when 2 levels are completed",
			pr:   &progress{CompletedLevels: &users.Enum[LevelType]{Level1Type, Level2Type}},
			expected: &Summary{Level: 2, Roles: []*Role{
				{Type: cfg.RoleNames[SnowmanRoleIndex], Enabled: true},
				{Type: cfg.RoleNames[AmbassadorRoleIndex], Enabled: false},
			}},
		},
	}

	for _, tt := range testCases { //nolint:gocritic // it's a test, no need for optimization
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := repo.newSummary(tt.pr, tt.requestingUserID)
			require.EqualValues(t, tt.expected, got)
		})
	}
}

func defaultCfg() *config {
	var cfg config
	const applicationYamlTestKey = applicationYamlKey + "_test"
	appcfg.MustLoadFromKey(applicationYamlTestKey, &cfg)

	return &cfg
}
