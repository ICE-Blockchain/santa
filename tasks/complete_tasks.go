// SPDX-License-Identifier: ice License 1.0

package tasks

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	friendsinvited "github.com/ice-blockchain/santa/friends-invited"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/log"
)

func (r *repository) PseudoCompleteTask(ctx context.Context, task *Task, dryRun bool) error { //nolint:funlen,revive,gocognit,gocyclo,cyclop // .
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	if err := r.validateTask(task); err != nil {
		return errors.Wrapf(err, "wrong type for:%+v", task)
	}
	userProgress, err := r.getProgress(ctx, task.UserID, true)
	if err != nil && !errors.Is(err, ErrRelationNotFound) {
		return errors.Wrapf(err, "failed to getProgress for userID:%v", task.UserID)
	}
	completed := userProgress.checkTaskCompleted(r, task)
	if r.cfg.TasksV2Enabled && dryRun {
		if !completed {
			return errors.Wrapf(ErrTaskNotCompleted, "task is not completed for: %v", task.UserID)
		}

		return nil
	}
	if r.cfg.TasksV2Enabled && userProgress != nil && !completed {
		return errors.Wrapf(ErrTaskNotCompleted, "task is not completed for: %v", task.UserID)
	}
	if userProgress == nil {
		userProgress = new(progress)
		userProgress.UserID = task.UserID
	}
	if task.Type == InviteFriendsType && userProgress.FriendsInvited < r.cfg.RequiredFriendsInvited {
		return nil
	}
	params, sql := userProgress.buildUpdatePseudoCompletedTasksSQL(task, r)
	if params == nil {
		// FE calls endpoint when task "completed", and we have nothing to update
		// so task is completed or pseudo-completed. But FE called endpoint one more time
		// it means there is some lag between pseudo and real-completion, may be msg is messed up,
		// so try to send another one to finally complete it (cuz it is already completed from FE side anyway).
		reallyCompleted := userProgress.reallyCompleted(task)
		if !reallyCompleted {
			return errors.Wrapf(r.sendTryCompleteTasksCommandMessage(ctx, task.UserID), "failed to sendTryCompleteTasksCommandMessage for userID:%v", task.UserID)
		}

		return nil
	}
	if updatedRows, uErr := storage.Exec(ctx, r.db, sql, params...); updatedRows == 0 && uErr == nil {
		return r.PseudoCompleteTask(ctx, task, dryRun)
	} else if uErr != nil {
		return errors.Wrapf(uErr, "failed to update task_progress.pseudo_completed_tasks for params:%#v", params...)
	}
	var sErr error
	if !r.cfg.TasksV2Enabled {
		sErr = r.sendTryCompleteTasksCommandMessage(ctx, task.UserID)
	} else {
		var prize float64
		for ix := range r.cfg.TasksList {
			if r.cfg.TasksList[ix].Type == string(task.Type) {
				prize = r.cfg.TasksList[ix].Prize
			}
		}
		sErr = r.sendCompletedTaskMessage(ctx, &CompletedTask{
			UserID: task.UserID,
			Type:   task.Type,
			Prize:  prize,
		})
	}
	if sErr != nil {
		sErr = errors.Wrapf(sErr, "failed to send message for userID:%v", task.UserID)
		pseudoCompletedTasks := params[2].(*users.Enum[Type]) //nolint:errcheck,forcetypeassert,revive // We're sure.
		sql = `UPDATE task_progress
		   SET pseudo_completed_tasks = $2
		   WHERE user_id = $1
			 AND COALESCE(pseudo_completed_tasks,ARRAY[]::TEXT[]) = COALESCE($3,ARRAY[]::TEXT[])`
		if updatedRows, rErr := storage.Exec(ctx, r.db, sql, task.UserID, userProgress.PseudoCompletedTasks, pseudoCompletedTasks); rErr == nil && updatedRows == 0 {
			log.Error(errors.Wrapf(sErr, "race condition while rolling back the update request"))

			return r.PseudoCompleteTask(ctx, task, dryRun)
		} else if rErr != nil {
			return multierror.Append( //nolint:wrapcheck // Not needed.
				sErr,
				errors.Wrapf(rErr, "[rollback] failed to update task_progress.pseudo_completed_tasks for params:%#v", params...),
			).ErrorOrNil()
		}

		return sErr
	}

	return nil
}

//nolint:funlen,gocyclo,revive,cyclop // .
func (p *progress) checkTaskCompleted(repo *repository, task *Task) bool {
	var completed bool
	switch task.Type { //nolint:exhaustive // Handling only v2 tasks here.
	case WatchVideoWithCodeConfirmation1Type:
		for ix := range repo.cfg.TasksList {
			if Type(repo.cfg.TasksList[ix].Type) == task.Type {
				if task.Data != nil {
					completed = task.Data.VerificationCode == repo.cfg.TasksList[ix].ConfirmationCode
				}

				break
			}
		}
	case InviteFriends5Type:
		completed = p.FriendsInvited >= 5
	case InviteFriends10Type:
		completed = p.FriendsInvited >= 10
	case InviteFriends25Type:
		completed = p.FriendsInvited >= 25
	case InviteFriends50Type:
		completed = p.FriendsInvited >= 50
	case InviteFriends100Type:
		completed = p.FriendsInvited >= 100
	case InviteFriends200Type:
		completed = p.FriendsInvited >= 200
	case ClaimCoinBadge1Type, ClaimCoinBadge2Type, ClaimCoinBadge3Type, ClaimCoinBadge4Type, ClaimCoinBadge5Type, ClaimCoinBadge6Type,
		ClaimCoinBadge7Type, ClaimCoinBadge8Type, ClaimCoinBadge9Type, ClaimCoinBadge10Type:
		parts := strings.Split(string(task.Type), "_")
		completed = p.isBadgeAchieved(parts[2])
	case ClaimLevelBadge1Type, ClaimLevelBadge2Type, ClaimLevelBadge3Type, ClaimLevelBadge4Type, ClaimLevelBadge5Type, ClaimLevelBadge6Type:
		parts := strings.Split(string(task.Type), "_")
		completed = p.isBadgeAchieved(parts[2])
	case ClaimSocialBadge1Type, ClaimSocialBadge2Type, ClaimSocialBadge3Type, ClaimSocialBadge4Type, ClaimSocialBadge5Type, ClaimSocialBadge6Type,
		ClaimSocialBadge7Type, ClaimSocialBadge8Type, ClaimSocialBadge9Type, ClaimSocialBadge10Type:
		parts := strings.Split(string(task.Type), "_")
		completed = p.isBadgeAchieved(parts[2])
	case ClaimLevel1Type, ClaimLevel2Type, ClaimLevel3Type, ClaimLevel4Type, ClaimLevel5Type, ClaimLevel6Type,
		ClaimLevel7Type, ClaimLevel8Type, ClaimLevel9Type, ClaimLevel10Type, ClaimLevel11Type, ClaimLevel12Type,
		ClaimLevel13Type, ClaimLevel14Type, ClaimLevel15Type, ClaimLevel16Type, ClaimLevel17Type, ClaimLevel18Type,
		ClaimLevel19Type, ClaimLevel20Type, ClaimLevel21Type:
		parts := strings.Split(string(task.Type), "_")
		completed = p.isLevelCompleted(parts[2])
	case MiningStreak7Type:
		completed = p.MiningStreak >= 7
	case MiningStreak14Type:
		completed = p.MiningStreak >= 14
	case MiningStreak30Type:
		completed = p.MiningStreak >= 30
	default:
		completed = true
	}

	return completed
}

//nolint:gocognit,nestif,revive // .
func (r *repository) validateTask(arg *Task) error {
	if r.cfg.TasksV2Enabled {
		for ix := range r.cfg.TasksList {
			if Type(r.cfg.TasksList[ix].Type) == arg.Type {
				return nil
			}
		}
	} else {
		for _, taskType := range &AllTypes {
			if taskType == arg.Type {
				if arg.Type == JoinTelegramType && (arg.Data == nil || arg.Data.TelegramUserHandle == "") {
					return errors.Wrap(ErrInvalidSocialProperties, "`data`.`telegramUserHandle` required")
				}
				if arg.Type == FollowUsOnTwitterType && (arg.Data == nil || arg.Data.TwitterUserHandle == "") {
					return errors.Wrap(ErrInvalidSocialProperties, "`data`.`twitterUserHandle` required")
				}

				return nil
			}
		}
	}

	return errors.Errorf("invalid type `%v`", arg.Type)
}

func (r *repository) tasksLength() int {
	if r.cfg.TasksV2Enabled {
		return len(r.cfg.TasksList)
	}

	return len(&AllTypes)
}

func (p *progress) buildUpdatePseudoCompletedTasksSQL(task *Task, repo *repository) (params []any, sql string) { //nolint:funlen // .
	for _, tsk := range p.buildTasks(repo) {
		if tsk.Type == task.Type {
			if tsk.Completed {
				return nil, ""
			}
		}
	}
	pseudoCompletedTasks := make(users.Enum[Type], 0, repo.tasksLength())
	if p.PseudoCompletedTasks != nil {
		pseudoCompletedTasks = append(pseudoCompletedTasks, *p.PseudoCompletedTasks...)
	}
	pseudoCompletedTasks = append(pseudoCompletedTasks, task.Type)
	if !repo.cfg.TasksV2Enabled {
		sort.SliceStable(pseudoCompletedTasks, func(i, j int) bool {
			return TypeOrder[pseudoCompletedTasks[i]] < TypeOrder[pseudoCompletedTasks[j]]
		})
	}
	params = make([]any, 0)
	params = append(params, task.UserID, p.PseudoCompletedTasks, &pseudoCompletedTasks)
	fieldIndexes := append(make([]string, 0, 1+1), "$3")
	fieldNames := append(make([]string, 0, 1+1), "pseudo_completed_tasks")
	nextIndex := 4
	switch task.Type { //nolint:exhaustive // We only care to treat those differently.
	case FollowUsOnTwitterType:
		params = append(params, task.Data.TwitterUserHandle)
		fieldNames = append(fieldNames, "twitter_user_handle")
		fieldIndexes = append(fieldIndexes, fmt.Sprintf("$%v ", nextIndex))
	case JoinTelegramType:
		if !repo.cfg.TasksV2Enabled {
			params = append(params, task.Data.TelegramUserHandle)
			fieldNames = append(fieldNames, "telegram_user_handle")
			fieldIndexes = append(fieldIndexes, fmt.Sprintf("$%v ", nextIndex))
		}
	}
	sql = fmt.Sprintf(`
			INSERT INTO task_progress(user_id, %[2]v)
			VALUES ($1, %[3]v)
			ON CONFLICT(user_id) DO UPDATE
									SET %[1]v
									WHERE COALESCE(task_progress.pseudo_completed_tasks,ARRAY[]::TEXT[]) = COALESCE($2,ARRAY[]::TEXT[])`,
		strings.Join(formatFields(fieldNames, fieldIndexes), ","),
		strings.Join(fieldNames, ","),
		strings.Join(fieldIndexes, ","),
	)

	return params, sql
}

func formatFields(names, indexes []string) []string {
	res := make([]string, len(names)) //nolint:makezero // We're know for sure.
	for ind, name := range names {
		res[ind] = fmt.Sprintf("%v = %v", name, indexes[ind])
	}

	return res
}

func (r *repository) completeTasks(ctx context.Context, userID string) error { //nolint:revive,funlen,gocognit,gocyclo,cyclop // .
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	pr, err := r.getProgress(ctx, userID, false)
	if err != nil && !errors.Is(err, ErrRelationNotFound) {
		return errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	if pr == nil {
		pr = new(progress)
		pr.UserID = userID
	}
	if pr.CompletedTasks != nil && len(*pr.CompletedTasks) == r.tasksLength() {
		return nil
	}
	completedTasks := pr.reEvaluateCompletedTasks(r)
	if completedTasks != nil && pr.CompletedTasks != nil {
		log.Debug(fmt.Sprintf("[completeTasks] for userID:%v progress:%#v(%v) completedTasks:%v", userID, *pr, *pr.CompletedTasks, *completedTasks))
	}
	if completedTasks != nil && pr.CompletedTasks != nil && len(*pr.CompletedTasks) == len(*completedTasks) {
		return nil
	}
	sql := `
		INSERT INTO task_progress(user_id, completed_tasks) VALUES ($1, $2)
			ON CONFLICT(user_id) DO UPDATE
				SET completed_tasks = EXCLUDED.completed_tasks
			WHERE COALESCE(task_progress.completed_tasks,ARRAY[]::TEXT[]) = COALESCE($3,ARRAY[]::TEXT[])`
	params := []any{
		pr.UserID,
		completedTasks,
		pr.CompletedTasks,
	}
	if updatedRows, uErr := storage.Exec(ctx, r.db, sql, params...); uErr == nil && updatedRows == 0 {
		return r.completeTasks(ctx, userID)
	} else if uErr != nil {
		return errors.Wrapf(uErr, "failed to insert/update task_progress.completed_tasks for params:%#v", params...)
	}
	//nolint:nestif // .
	if completedTasks != nil && len(*completedTasks) > 0 && (pr.CompletedTasks == nil || len(*pr.CompletedTasks) < len(*completedTasks)) {
		newlyCompletedTasks := make([]*CompletedTask, 0, r.tasksLength())
	outer:
		for _, completedTask := range *completedTasks {
			if pr.CompletedTasks != nil {
				for _, previouslyCompletedTask := range *pr.CompletedTasks {
					if completedTask == previouslyCompletedTask {
						continue outer
					}
				}
			}
			if !r.cfg.TasksV2Enabled {
				newlyCompletedTasks = append(newlyCompletedTasks, &CompletedTask{
					UserID:         userID,
					Type:           completedTask,
					CompletedTasks: uint64(len(*completedTasks)),
				})
			}
		}
		if err = runConcurrently(ctx, r.sendCompletedTaskMessage, newlyCompletedTasks); err != nil {
			sErr := errors.Wrapf(err, "failed to sendCompletedTaskMessages for userID:%v,completedTasks:%#v", userID, newlyCompletedTasks)
			params[1] = pr.CompletedTasks
			params[2] = completedTasks
			if updatedRows, rErr := storage.Exec(ctx, r.db, sql, params...); rErr == nil && updatedRows == 0 {
				log.Error(errors.Wrapf(sErr, "[sendCompletedTaskMessages]rollback race condition"))

				return r.completeTasks(ctx, userID)
			} else if rErr != nil {
				return multierror.Append( //nolint:wrapcheck // Not needed.
					sErr,
					errors.Wrapf(rErr, "[sendCompletedTaskMessages][rollback] failed to update task_progress.completed_tasks for params:%#v", params...),
				).ErrorOrNil()
			}

			return sErr
		}
	}

	return nil
}

func (p *progress) reEvaluateCompletedTasks(repo *repository) *users.Enum[Type] { //nolint:revive,funlen,gocognit // .
	if p.CompletedTasks != nil && len(*p.CompletedTasks) == repo.tasksLength() {
		return p.CompletedTasks
	}
	alreadyCompletedTasks := make(map[Type]any, repo.tasksLength())
	if p.CompletedTasks != nil {
		for _, task := range *p.CompletedTasks {
			alreadyCompletedTasks[task] = struct{}{}
		}
	}
	completedTasks := make(users.Enum[Type], 0, repo.tasksLength())
	if repo.cfg.TasksV2Enabled { //nolint:nestif // .
		for ix := range repo.cfg.TasksList {
			if _, alreadyCompleted := alreadyCompletedTasks[Type(repo.cfg.TasksList[ix].Type)]; alreadyCompleted {
				completedTasks = append(completedTasks, Type(repo.cfg.TasksList[ix].Type))

				continue
			}
			if val := p.gatherCompletedTasks(repo, Type(repo.cfg.TasksList[ix].Type)); val != "" {
				completedTasks = append(completedTasks, val)
			}
		}
	} else {
		for _, taskType := range &AllTypes {
			if _, alreadyCompleted := alreadyCompletedTasks[taskType]; alreadyCompleted {
				completedTasks = append(completedTasks, taskType)

				continue
			}
			if val := p.gatherCompletedTasks(repo, taskType); val != "" {
				completedTasks = append(completedTasks, val)
			}
		}
	}
	if len(completedTasks) == 0 {
		return nil
	}

	return &completedTasks
}

func (p *progress) gatherCompletedTasks(repo *repository, taskType Type) Type {
	var completed bool
	switch taskType { //nolint:exhaustive // Handling v1 tasks mostly, v2 are go through pseudo tasks.
	case ClaimUsernameType:
		completed = p.UsernameSet
	case StartMiningType:
		completed = p.MiningStarted
	case UploadProfilePictureType:
		completed = p.ProfilePictureSet
	case JoinTwitterType:
		completed = p.checkPseudoTaskCompleted(taskType)
	case FollowUsOnTwitterType:
		completed = p.TwitterUserHandle != nil && *p.TwitterUserHandle != ""
	case JoinTelegramType:
		if repo.cfg.TasksV2Enabled {
			completed = p.checkPseudoTaskCompleted(taskType)
		} else if p.TelegramUserHandle != nil && *p.TelegramUserHandle != "" {
			completed = true
		}
	case InviteFriendsType:
		completed = p.FriendsInvited >= repo.cfg.RequiredFriendsInvited
	default:
		completed = p.checkPseudoTaskCompleted(taskType)
	}
	if completed {
		return taskType
	}

	return ""
}

func (p *progress) isBadgeAchieved(badge string) bool {
	if p.AchievedBadges != nil {
		for _, achievedBadge := range *p.AchievedBadges {
			if achievedBadge == badge {
				return true
			}
		}
	}

	return false
}

func (p *progress) isLevelCompleted(level string) bool {
	if p.CompletedLevels != nil {
		for _, completedLevel := range *p.CompletedLevels {
			if completedLevel == level {
				return true
			}
		}
	}

	return false
}

func (p *progress) checkPseudoTaskCompleted(taskType Type) bool {
	if p.PseudoCompletedTasks != nil {
		for ix := range *p.PseudoCompletedTasks {
			if (*p.PseudoCompletedTasks)[ix] == taskType {
				return true
			}
		}
	}

	return false
}

func (r *repository) sendCompletedTaskMessage(ctx context.Context, completedTask *CompletedTask) error {
	valueBytes, err := json.MarshalContext(ctx, completedTask)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal %#v", completedTask)
	}
	msg := &messagebroker.Message{
		Headers: map[string]string{"producer": "santa"},
		Key:     completedTask.UserID,
		Topic:   r.cfg.MessageBroker.Topics[2].Name,
		Value:   valueBytes,
	}
	responder := make(chan error, 1)
	defer close(responder)
	r.mb.SendMessage(ctx, msg, responder)

	return errors.Wrapf(<-responder, "failed to send `%v` message to broker", msg.Topic)
}

func (r *repository) sendTryCompleteTasksCommandMessage(ctx context.Context, userID string) error {
	msg := &messagebroker.Message{
		Headers: map[string]string{"producer": "santa"},
		Key:     userID,
		Topic:   r.cfg.MessageBroker.Topics[1].Name,
	}
	responder := make(chan error, 1)
	defer close(responder)
	r.mb.SendMessage(ctx, msg, responder)

	return errors.Wrapf(<-responder, "failed to send `%v` message to broker", msg.Topic)
}

func (s *tryCompleteTasksCommandSource) Process(ctx context.Context, msg *messagebroker.Message) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline while processing message")
	}
	if !s.cfg.TasksV2Enabled {
		return errors.Wrapf(s.completeTasks(ctx, msg.Key), "failed to completeTasks for userID:%v", msg.Key)
	}

	return nil
}

func (s *miningSessionSource) Process(ctx context.Context, msg *messagebroker.Message) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline while processing message")
	}
	if len(msg.Value) == 0 {
		return nil
	}
	var ses struct {
		UserID string `json:"userId,omitempty" example:"did:ethr:0x4B73C58370AEfcEf86A6021afCDe5673511376B2"`
	}
	if err := json.UnmarshalContext(ctx, msg.Value, &ses); err != nil {
		return errors.Wrapf(err, "process: cannot unmarshall %v into %#v", string(msg.Value), &ses)
	}
	if ses.UserID == "" {
		return nil
	}

	return errors.Wrapf(s.upsertProgress(ctx, ses.UserID), "failed to upsertProgress for userID:%v", ses.UserID)
}

func (s *miningSessionSource) upsertProgress(ctx context.Context, userID string) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "context failed")
	}
	if pr, err := s.getProgress(ctx, userID, true); (pr != nil && pr.CompletedTasks != nil &&
		len(*pr.CompletedTasks) == s.tasksLength()) || err != nil && !errors.Is(err, ErrRelationNotFound) {
		return errors.Wrapf(err, "failed to getProgress for userID:%v", userID)
	}
	sql := `INSERT INTO task_progress(user_id, mining_started) VALUES ($1, $2)
			ON CONFLICT(user_id) DO UPDATE SET mining_started = EXCLUDED.mining_started
			WHERE task_progress.mining_started != EXCLUDED.mining_started;`
	_, err := storage.Exec(ctx, s.db, sql, userID, true)

	return multierror.Append( //nolint:wrapcheck // Not needed.
		errors.Wrapf(err, "failed to upsert progress for %v %v", userID, true),
		errors.Wrapf(s.sendTryCompleteTasksCommandMessage(ctx, userID),
			"failed to sendTryCompleteTasksCommandMessage for userID:%v", userID),
	).ErrorOrNil()
}

func (s *userTableSource) Process(ctx context.Context, msg *messagebroker.Message) error {
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
	if (snapshot.Before == nil || snapshot.Before.ID == "") && (snapshot.User == nil || snapshot.User.ID == "") {
		return nil
	}
	if snapshot.Before != nil && snapshot.Before.ID != "" && (snapshot.User == nil || snapshot.User.ID == "") {
		return errors.Wrapf(s.deleteProgress(ctx, snapshot), "failed to delete progress for:%#v", snapshot)
	}

	return errors.Wrapf(s.upsertProgress(ctx, snapshot), "failed to upsert progress for:%#v", snapshot)
}

func (s *userTableSource) upsertProgress(ctx context.Context, us *users.UserSnapshot) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "context failed")
	}
	insertTuple := &progress{
		UserID:            us.ID,
		UsernameSet:       us.Username != "" && us.Username != us.ID,
		ProfilePictureSet: us.ProfilePictureURL != "" && !strings.Contains(us.ProfilePictureURL, "default-profile-picture"),
	}
	sql := `INSERT INTO task_progress(user_id, username_set, profile_picture_set) VALUES ($1, $2, $3)
			ON CONFLICT(user_id) DO UPDATE SET 
			        username_set = EXCLUDED.username_set,
			    	profile_picture_set = EXCLUDED.profile_picture_set
			WHERE task_progress.username_set != EXCLUDED.username_set
			   OR task_progress.profile_picture_set != EXCLUDED.profile_picture_set;`
	_, err := storage.Exec(ctx, s.db, sql, insertTuple.UserID, insertTuple.UsernameSet, insertTuple.ProfilePictureSet)

	return multierror.Append( //nolint:wrapcheck // Not needed.
		errors.Wrapf(err, "failed to upsert progress for %#v", us),
		errors.Wrapf(s.sendTryCompleteTasksCommandMessage(ctx, us.ID), "failed to sendTryCompleteTasksCommandMessage for userID:%v", us.ID),
	).ErrorOrNil()
}

func (f *friendsInvitedSource) Process(ctx context.Context, msg *messagebroker.Message) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline while processing message")
	}
	if len(msg.Value) == 0 {
		return nil
	}
	friends := new(friendsinvited.Count)
	if err := json.UnmarshalContext(ctx, msg.Value, friends); err != nil {
		return errors.Wrapf(err, "cannot unmarshal %v into %#v", string(msg.Value), friends)
	}

	return multierror.Append( //nolint:wrapcheck // Not needed.
		errors.Wrapf(f.updateFriendsInvited(ctx, friends), "failed to update tasks friends invited for %#v", friends),
		errors.Wrapf(f.sendTryCompleteTasksCommandMessage(ctx, friends.UserID), "failed to sendTryCompleteTasksCommandMessage for userID:%v", friends.UserID),
	).ErrorOrNil()
}

func (f *friendsInvitedSource) updateFriendsInvited(ctx context.Context, friends *friendsinvited.Count) error {
	sql := `INSERT INTO task_progress(user_id, friends_invited) VALUES ($1, $2)
		   ON CONFLICT(user_id) DO UPDATE  
		   		SET friends_invited = EXCLUDED.friends_invited
		   	WHERE task_progress.friends_invited != EXCLUDED.friends_invited`
	_, err := storage.Exec(ctx, f.db, sql, friends.UserID, friends.FriendsInvited)

	return errors.Wrapf(err, "failed to set task_progress.friends_invited, params:%#v", friends)
}

func (s *userTableSource) deleteProgress(ctx context.Context, us *users.UserSnapshot) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "context failed")
	}
	params := []any{us.Before.ID}
	_, errDelUser := storage.Exec(ctx, s.db, `DELETE FROM task_progress WHERE user_id = $1`, params...)

	return errors.Wrapf(errDelUser, "failed to delete task_progress for:%#v", us)
}
