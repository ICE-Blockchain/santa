// SPDX-License-Identifier: ice License 1.0

package main

import (
	"context"
	_ "embed"
	"sync"

	"github.com/pkg/errors"

	storagepg "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/log"
)

const (
	applicationYamlUsersKey = "users"
	applicationYamlKeySanta = "santa"
	concurrencyCount        = 1000
)

type (
	updater struct {
		dbSanta  *storagepg.DB
		dbEskimo *storagepg.DB
	}
	eskimoUser struct {
		ID             string `json:"id" example:"did:ethr:0x4B73C58370AEfcEf86A6021afCDe5673511376B2" db:"id"`
		FriendsInvited uint64 `json:"friendsInvited" example:"22" db:"friends_invited"`
	}

	commonUser struct {
		UserID string
	}
)

func main() {
	dbEskimo := storagepg.MustConnect(context.Background(), "", applicationYamlUsersKey)
	dbSanta := storagepg.MustConnect(context.Background(), "", applicationYamlKeySanta)

	if err := dbEskimo.Ping(context.Background()); err != nil {
		log.Panic("can't ping users db", err)
	}
	if err := dbSanta.Ping(context.Background()); err != nil {
		log.Panic("can't ping santa db", err)
	}
	upd := &updater{
		dbSanta:  dbSanta,
		dbEskimo: dbEskimo,
	}
	defer upd.dbEskimo.Close()
	defer upd.dbSanta.Close()

	upd.update(context.Background())
}

//nolint:revive,funlen,gocognit // .
func (u *updater) update(ctx context.Context) {
	var (
		updatedCount uint64
		maxLimit     uint64 = 10000
		offset       uint64
	)
	concurrencyGuard := make(chan struct{}, concurrencyCount)
	wg := new(sync.WaitGroup)
	for {
		/******************************************************************************************************************************************************
			1. Fetching a new batch of users from eskimo.
		******************************************************************************************************************************************************/
		sql := `SELECT 
					u.id,
					COUNT(DISTINCT t1.id) 		AS friends_invited
				FROM users u
				LEFT JOIN USERS t1
					ON t1.referred_by = u.ID
						AND t1.id != u.id
						AND t1.username != t1.id
						AND t1.referred_by != t1.id
				JOIN referral_acquisition_history rah
					ON u.id = rah.user_id
				GROUP BY u.id, rah.t1
				ORDER BY u.created_at ASC
				LIMIT $1
				OFFSET $2`
		usrs, err := storagepg.Select[eskimoUser](ctx, u.dbEskimo, sql, maxLimit, offset)
		if err != nil {
			log.Panic("error on trying to get actual friends invited values crossed with already updated values", err)
		}
		if len(usrs) == 0 {
			break
		}

		/******************************************************************************************************************************************************
			2. Fetching tasks progress data.
		******************************************************************************************************************************************************/
		var userKeysProgress []string
		actualFriendsInvitedCount := make(map[string]uint64, len(usrs))
		for _, usr := range usrs {
			if usr.ID == "" {
				continue
			}
			userKeysProgress = append(userKeysProgress, usr.ID)
			actualFriendsInvitedCount[usr.ID] = usr.FriendsInvited
		}
		sql = `SELECT 
					tp.user_id
				FROM task_progress tp
					WHERE tp.user_id = ANY($1)`
		res, err := storagepg.Select[commonUser](ctx, u.dbSanta, sql, userKeysProgress)
		if err != nil {
			log.Panic("error on trying to get tasks", userKeysProgress, err)
		}
		if len(res) == 0 {
			offset += maxLimit

			continue
		}

		/******************************************************************************************************************************************************
			3. Updating santa.
		******************************************************************************************************************************************************/
		for _, r := range res {
			if r.UserID == "" {
				continue
			}
			usr := r
			wg.Add(1)
			concurrencyGuard <- struct{}{}
			go func() {
				defer wg.Done()
				if bErr := u.updateBadges(ctx, usr, actualFriendsInvitedCount[usr.UserID]); bErr != nil {
					log.Panic("can't update badges, userID:", usr.UserID, bErr)
				}
				if uErr := u.updateTasks(ctx, usr, actualFriendsInvitedCount[usr.UserID]); uErr != nil {
					log.Panic("can't update tasks, userID:", usr.UserID, uErr)
				}
				if uErr := u.updateFriendsInvited(ctx, usr, actualFriendsInvitedCount[usr.UserID]); uErr != nil {
					log.Panic("can't update friends invited, userID:", usr.UserID, uErr)
				}
				<-concurrencyGuard
			}()
		}

		updatedCount += uint64(len(res))
		log.Info("updated count: ", updatedCount)

		offset += maxLimit
	}
	wg.Wait()
}

func (u *updater) updateBadges(ctx context.Context, usr *commonUser, actualFriendsInvited uint64) error {
	sql := `UPDATE badge_progress
								SET friends_invited = $2
							WHERE user_id = $1
								  AND friends_invited != $2`
	_, err := storagepg.Exec(ctx, u.dbSanta, sql, usr.UserID, actualFriendsInvited)

	return errors.Wrapf(err, "failed to update badge_progress, userID:%v, friendsInvited:%v", usr.UserID, actualFriendsInvited)
}

func (u *updater) updateTasks(ctx context.Context, usr *commonUser, actualFriendsInvited uint64) error {
	sql := `UPDATE task_progress
								SET friends_invited = $2
							WHERE user_id = $1
								  AND (friends_invited != $2)`
	_, err := storagepg.Exec(ctx, u.dbSanta, sql, usr.UserID, actualFriendsInvited)

	return errors.Wrapf(err, "failed to update task_progress, userID:%v, friendsInvited:%v", usr.UserID, actualFriendsInvited)
}

func (u *updater) updateFriendsInvited(ctx context.Context, usr *commonUser, actualFriendsInvited uint64) error {
	sql := `UPDATE friends_invited
				   SET invited_count = $2
				   WHERE user_id = $1
				   		 AND friends_invited.invited_count != $2`
	_, err := storagepg.Exec(ctx, u.dbSanta, sql, usr.UserID, actualFriendsInvited)

	return errors.Wrapf(err, "failed to update friends invited, userID:%v, friendsInvited:%v", usr.UserID, actualFriendsInvited)
}
