// SPDX-License-Identifier: ice License 1.0

package tasks

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
)

//nolint:funlen,gocognit,revive,gocognit,gocyclo,cyclop // .
func (r *repository) GetTasks(ctx context.Context, userID, language string, requestedStatus TaskStatus) (resp []*Task, err error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	userProgress, err := r.getProgress(ctx, userID, true)
	if err != nil {
		if errors.Is(err, ErrRelationNotFound) {
			return r.defaultTasks(), nil
		}

		return nil, errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	tasks := userProgress.buildTasks(r)
	if r.cfg.TasksV2Enabled { //nolint:nestif // .
		if requestedStatus != TaskStatusCompleted && requestedStatus != TaskStatusPending {
			return nil, errors.Wrapf(ErrWrongRequestedTasksStatus, "requested status should be:%v or %v", TaskStatusCompleted, TaskStatusPending)
		}
		lang := language
		if language == "" {
			lang = defaultLanguage
		}
		taskGroups := make(map[string]*Task)
		for _, task := range tasks {
			if (requestedStatus == TaskStatusPending && task.Completed) ||
				(requestedStatus == TaskStatusCompleted && !task.Completed) {
				continue
			}
			task.prepareTranslations(lang, r.cfg)
			if requestedStatus == TaskStatusPending && (task.Group != "") {
				splitted := strings.Split(string(task.Type), "_")
				if (task.Group == TaskGroupBadgeSocial || task.Group == TaskGroupBadgeCoin || task.Group == TaskGroupBadgeLevel) &&
					!userProgress.isBadgeAchieved(splitted[2]) {
					continue
				}
				if task.Group == TaskGroupLevel && !userProgress.isLevelCompleted(splitted[2]) {
					continue
				}
				if _, ok := taskGroups[task.Group]; !ok && !task.Completed {
					taskGroups[task.Group] = task
					resp = append(resp, task)
				}

				continue
			}

			resp = append(resp, task)
		}
		if len(resp) == 0 {
			return []*Task{}, nil
		}
		sanitizeTasksForUI(resp)

		return resp, nil
	}

	return tasks, nil
}

func sanitizeTasksForUI(tasks []*Task) {
	for ix := range tasks {
		tasks[ix].Metadata.LongDescription = ""
		tasks[ix].Metadata.ErrorDescription = ""
	}
}

//nolint:revive //.
func (r *repository) getProgress(ctx context.Context, userID string, tolerateOldData bool) (res *progress, err error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	sql := `SELECT task_progress.*, levels_and_roles_progress.completed_levels, badge_progress.achieved_badges, levels_and_roles_progress.mining_streak
			FROM task_progress
			JOIN levels_and_roles_progress
				ON levels_and_roles_progress.user_id = task_progress.user_id
			JOIN badge_progress
				ON badge_progress.user_id = task_progress.user_id
			WHERE task_progress.user_id = $1`
	if tolerateOldData {
		res, err = storage.Get[progress](ctx, r.db, sql, userID)
	} else {
		res, err = storage.ExecOne[progress](ctx, r.db, sql, userID)
	}

	err = errors.Wrapf(err, "failed to get TASK_PROGRESS for userID:%v", userID)
	if storage.IsErr(err, storage.ErrNotFound) {
		return nil, ErrRelationNotFound
	}

	return
}

func (p *progress) buildTasks(repo *repository) []*Task { //nolint:gocognit,funlen,revive // Wrong.
	resp := repo.defaultTasks()
	for _, task := range resp {
		switch task.Type { //nolint:exhaustive // Only those 2 have specific data persisted.
		case JoinTwitterType, FollowUsOnTwitterType:
			if p.TwitterUserHandle != nil && *p.TwitterUserHandle != "" {
				task.Data = &Data{
					TwitterUserHandle: *p.TwitterUserHandle,
				}
			}
		case JoinTelegramType:
			if p.TelegramUserHandle != nil && *p.TelegramUserHandle != "" {
				task.Data = &Data{
					TelegramUserHandle: *p.TelegramUserHandle,
				}
			}
		}
		if p.CompletedTasks != nil {
			for _, completedTask := range *p.CompletedTasks {
				if task.Type == completedTask {
					task.Completed = true

					break
				}
			}
		}
		if p.PseudoCompletedTasks != nil && !task.Completed {
			for _, pseudoCompletedTask := range *p.PseudoCompletedTasks {
				if task.Type == pseudoCompletedTask {
					task.Completed = true

					break
				}
			}
		}
	}

	return resp
}

func (p *progress) reallyCompleted(task *Task) bool {
	if p.CompletedTasks == nil {
		return false
	}
	reallyCompleted := false
	for _, tsk := range *p.CompletedTasks {
		if tsk == task.Type {
			reallyCompleted = true

			break
		}
	}

	return reallyCompleted
}

func (r *repository) defaultTasks() (resp []*Task) {
	if r.cfg.TasksV2Enabled {
		return r.defaultTasksV2()
	}

	return r.defaultTasksV1()
}

func (r *repository) defaultTasksV1() (resp []*Task) {
	resp = make([]*Task, 0, len(&AllTypes))
	for _, taskType := range &AllTypes {
		var (
			data      *Data
			completed bool
		)
		switch taskType { //nolint:exhaustive // We care only about those.
		case ClaimUsernameType:
			completed = true // To make sure network latency doesn't affect UX.
		case InviteFriendsType:
			data = &Data{RequiredQuantity: r.cfg.RequiredFriendsInvited}
		}
		resp = append(resp, &Task{Data: data, Type: taskType, Completed: completed})
	}

	return resp
}

func (r *repository) defaultTasksV2() (resp []*Task) {
	resp = make([]*Task, 0, len(r.cfg.TasksList))
	for ix := range r.cfg.TasksList {
		var (
			data      *Data
			completed bool
		)
		if r.cfg.TasksList[ix].Group == TaskGroupInviteFriends || r.cfg.TasksList[ix].Group == TaskGroupMiningStreak {
			data = &Data{RequiredQuantity: r.cfg.TasksList[ix].RequiredQuantity}
		}
		task := &Task{
			Data:       data,
			Type:       Type(r.cfg.TasksList[ix].Type),
			Completed:  completed,
			Group:      r.cfg.TasksList[ix].Group,
			Prize:      r.cfg.TasksList[ix].Prize,
			GroupIndex: r.cfg.TasksList[ix].GroupIndex,
		}
		task.Metadata = &Metadata{
			IconURL: r.cfg.TasksList[ix].Icon,
		}
		if r.cfg.TasksList[ix].URL != "" {
			task.Metadata.TaskURL = r.cfg.TasksList[ix].URL
		}
		resp = append(resp, task)
	}

	return resp
}

//nolint:funlen,gocognit,revive // .
func (r *repository) GetTask(ctx context.Context, userID, language string, taskType Type) (resp *Task, err error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	if !r.cfg.TasksV2Enabled {
		return nil, ErrNotSupported
	}
	userProgress, err := r.getProgress(ctx, userID, true)
	if err != nil {
		if errors.Is(err, ErrRelationNotFound) {
			for _, task := range r.defaultTasks() {
				if task.Type == taskType {
					return task, nil
				}
			}

			return nil, err
		}

		return nil, errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	tasks := userProgress.buildTasks(r)
	lang := language
	if language == "" {
		lang = defaultLanguage
	}
	for _, task := range tasks {
		if task.Type != taskType {
			continue
		}
		task.prepareTranslations(lang, r.cfg)

		return task, nil
	}

	return nil, ErrNotFound
}

//nolint:funlen // .
func (t *Task) prepareTranslations(language string, cfg *config) {
	tmpl := allTaskTemplates[t.Type][language]
	if tm, ok := allTaskTemplates[t.Type][language]; !ok || tm == nil {
		tmpl = allTaskTemplates[t.Type][defaultLanguage]
	}
	switch t.Group {
	case TaskGroupBadgeCoin:
		t.Metadata.Title = tmpl.getTitle(struct{ BadgeName string }{
			BadgeName: cfg.BadgesList.Coins[t.GroupIndex-1].Name,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(nil)
		t.Metadata.LongDescription = tmpl.getLongDescription(nil)
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(nil)
	case TaskGroupBadgeLevel:
		t.Metadata.Title = tmpl.getTitle(struct{ BadgeName string }{
			BadgeName: cfg.BadgesList.Levels[t.GroupIndex-1].Name,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(nil)
		t.Metadata.LongDescription = tmpl.getLongDescription(nil)
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(nil)
	case TaskGroupBadgeSocial:
		t.Metadata.Title = tmpl.getTitle(struct{ BadgeName string }{
			BadgeName: cfg.BadgesList.Socials[t.GroupIndex-1].Name,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(nil)
		t.Metadata.LongDescription = tmpl.getLongDescription(nil)
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(nil)
	case TaskGroupLevel:
		t.Metadata.Title = tmpl.getTitle(struct{ LevelNumber uint64 }{
			LevelNumber: t.GroupIndex,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(nil)
		t.Metadata.LongDescription = tmpl.getLongDescription(nil)
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(nil)
	case TaskGroupInviteFriends:
		t.Metadata.Title = tmpl.getTitle(struct{ RequiredFriendsInvitedCount uint64 }{
			RequiredFriendsInvitedCount: t.Data.RequiredQuantity,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(struct{ RequiredFriendsInvitedCount uint64 }{
			RequiredFriendsInvitedCount: t.Data.RequiredQuantity,
		})
		t.Metadata.LongDescription = tmpl.getLongDescription(struct{ RequiredFriendsInvitedCount uint64 }{
			RequiredFriendsInvitedCount: t.Data.RequiredQuantity,
		})
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(struct{ RequiredFriendsInvitedCount uint64 }{
			RequiredFriendsInvitedCount: t.Data.RequiredQuantity,
		})
	case TaskGroupMiningStreak:
		t.Metadata.Title = tmpl.getTitle(struct{ RequiredMiningStreakNumber uint64 }{
			RequiredMiningStreakNumber: t.Data.RequiredQuantity,
		})
		t.Metadata.ShortDescription = tmpl.getShortDescription(struct{ RequiredMiningStreakNumber uint64 }{
			RequiredMiningStreakNumber: t.Data.RequiredQuantity,
		})
		t.Metadata.LongDescription = tmpl.getLongDescription(struct{ RequiredMiningStreakNumber uint64 }{
			RequiredMiningStreakNumber: t.Data.RequiredQuantity,
		})
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(struct{ RequiredMiningStreakNumber uint64 }{
			RequiredMiningStreakNumber: t.Data.RequiredQuantity,
		})
	default:
		t.Metadata.Title = tmpl.getTitle(nil)
		t.Metadata.ShortDescription = tmpl.getShortDescription(nil)
		t.Metadata.LongDescription = tmpl.getLongDescription(nil)
		t.Metadata.ErrorDescription = tmpl.getErrorDescription(nil)
	}
}
