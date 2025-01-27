// SPDX-License-Identifier: ice License 1.0

package friendsinvited

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	"github.com/ice-blockchain/wintr/connectors/storage/v2"
)

func (s *userTableSource) Process(ctx context.Context, msg *messagebroker.Message) error { //nolint:gocyclo,revive,cyclop // .
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline while processing message")
	}
	if len(msg.Value) == 0 {
		return nil
	}
	snapshot := new(users.UserSnapshot)
	if err := json.UnmarshalContext(ctx, msg.Value, snapshot); err != nil {
		return errors.Wrapf(err, "cannot unmarshal %v into %#v", string(msg.Value), snapshot)
	}
	referredByChangedOnMofidy := snapshot.Before != nil && snapshot.Before.ID != "" && snapshot.User != nil && snapshot.User.ID != "" &&
		snapshot.User.ReferredBy != snapshot.User.ID && snapshot.User.ReferredBy != "" &&
		(snapshot.Before.ReferredBy == snapshot.Before.ID || snapshot.Before.ReferredBy == "")
	referredByChangedOnCreate := snapshot.Before == nil && snapshot.User != nil && snapshot.User.ID != "" &&
		snapshot.User.ReferredBy != snapshot.User.ID && snapshot.User.ReferredBy != ""
	if referredByChangedOnCreate || referredByChangedOnMofidy {
		return errors.Wrapf(s.insertReferrals(ctx, snapshot), "failed to insertReferrals[friendsinvited] for:%#v", snapshot)
	}
	if snapshot.Before != nil && snapshot.Before.ID != "" && (snapshot.User == nil || snapshot.User.ID == "") {
		return errors.Wrapf(s.deleteFriendsInvited(ctx, snapshot), "failed to delete [friendsinvited] progress for:%#v", snapshot)
	}

	return nil
}

func (s *userTableSource) insertReferrals(ctx context.Context, us *users.UserSnapshot) error {
	return errors.Wrapf(storage.DoInTransaction(ctx, s.db, func(tx storage.QueryExecer) error {
		sql := `INSERT INTO referrals(user_id,referred_by, processed_at, deleted) VALUES ($1,$2,$3, false)`
		params := []any{
			us.User.ID,
			us.User.ReferredBy,
			us.User.UpdatedAt.Time,
		}
		if _, err := storage.Exec(ctx, tx, sql, params...); err != nil {
			if storage.IsErr(err, storage.ErrDuplicate) {
				return nil
			}

			return errors.Wrapf(err, "failed to insert referrals, params:%#v", params...)
		}
		sql = `
		INSERT INTO friends_invited(user_id,invited_count) VALUES ($1, 1)
		ON CONFLICT(user_id) DO UPDATE SET
		   invited_count = friends_invited.invited_count + 1
		RETURNING *`
		friends, err := storage.ExecOne[Count](ctx, tx, sql, us.User.ReferredBy)
		if err != nil {
			return errors.Wrapf(err, "failed to increment friends_invited for userID:%v (ref:%v)", us.User.ReferredBy, us.User.ID)
		}

		return s.sendFriendsInvitedCountUpdate(ctx, friends)
	}), "insertReferrals: transaction failed for %#v", us)
}

func (s *userTableSource) deleteFriendsInvited(ctx context.Context, us *users.UserSnapshot) error { //nolint:funlen // .
	return errors.Wrapf(storage.DoInTransaction(ctx, s.db, func(tx storage.QueryExecer) error {
		if _, errDelUser := storage.Exec(ctx, tx, `DELETE FROM friends_invited WHERE user_id = $1`, us.Before.ID); errDelUser != nil {
			return errors.Wrapf(errDelUser, "failed to delete friends-invited for:%#v", us)
		}
		if us.Before.ReferredBy == "" || us.Before.ReferredBy == us.Before.ID {
			return nil
		}
		sql := `INSERT INTO referrals(user_id,referred_by, processed_at, deleted) VALUES ($1,$2,$3, true)`
		params := []any{us.Before.ID, us.Before.ReferredBy, us.Before.UpdatedAt.Time}
		if _, err := storage.Exec(ctx, tx, sql, params...); err != nil {
			if storage.IsErr(err, storage.ErrDuplicate) {
				return nil
			}

			return errors.Wrapf(err, "failed to insert referrals, params:%#v", params...)
		}
		updatedFriendsCount, errDecrementT0 := storage.ExecOne[Count](ctx, tx, `
			UPDATE friends_invited SET
				invited_count = GREATEST(friends_invited.invited_count - 1, 0)
			WHERE user_id = $1
			RETURNING *`, us.Before.ReferredBy)
		if errDecrementT0 != nil {
			if storage.IsErr(errDecrementT0, storage.ErrNotFound) {
				return nil
			}

			return errors.Wrapf(errDecrementT0, "failed to decrement friends-invited for T0:%v due to user %v deletion", us.Before.ReferredBy, us.Before.ID)
		}

		return errors.Wrapf(s.sendFriendsInvitedCountUpdate(ctx, updatedFriendsCount),
			"failed to sendFriendsInvitedCountUpdate for T0:%v due to user %v deletion", us.Before.ReferredBy, us.Before.ID)
	}), "deleteFriendsInvited: transaction failed for %#v", us)
}

func (r *repository) sendFriendsInvitedCountUpdate(ctx context.Context, friends *Count) error {
	valueBytes, err := json.MarshalContext(ctx, friends)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal %#v", friends)
	}
	msg := &messagebroker.Message{
		Headers: map[string]string{"producer": "santa"},
		Key:     friends.UserID,
		Topic:   r.cfg.MessageBroker.Topics[1].Name,
		Value:   valueBytes,
	}
	responder := make(chan error, 1)
	defer close(responder)
	r.mb.SendMessage(ctx, msg, responder)

	return errors.Wrapf(<-responder, "failed to send `%v` message to broker", msg.Topic)
}
