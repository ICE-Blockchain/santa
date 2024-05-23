// SPDX-License-Identifier: ice License 1.0

package badges

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	appcfg "github.com/ice-blockchain/wintr/config"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/time"
)

func New(ctx context.Context, _ context.CancelFunc) Repository {
	var cfg config
	appcfg.MustLoadFromKey(applicationYamlKey, &cfg)

	db := storage.MustConnect(ctx, ddl, applicationYamlKey)
	loadBadges(&cfg)

	return &repository{
		cfg:      &cfg,
		shutdown: db.Close,
		db:       db,
	}
}

func StartProcessor(ctx context.Context, cancel context.CancelFunc) Processor {
	var cfg config
	appcfg.MustLoadFromKey(applicationYamlKey, &cfg)
	loadBadges(&cfg)

	var mbConsumer messagebroker.Client
	prc := &processor{repository: &repository{
		cfg: &cfg,
		db:  storage.MustConnect(ctx, ddl, applicationYamlKey),
		mb:  messagebroker.MustConnect(ctx, applicationYamlKey),
	}}
	mbConsumer = messagebroker.MustConnectAndStartConsuming(context.Background(), cancel, applicationYamlKey, //nolint:contextcheck // .
		&tryAchievedBadgesCommandSource{processor: prc},
		&achievedBadgesSource{processor: prc},
		&userTableSource{processor: prc},
		&completedLevelsSource{processor: prc},
		&balancesTableSource{processor: prc},
		&globalTableSource{processor: prc},
		&friendsInvitedSource{processor: prc},
	)
	prc.shutdown = closeAll(mbConsumer, prc.mb, prc.db)

	return prc
}

func (r *repository) Close() error {
	return errors.Wrap(r.shutdown(), "closing repository failed")
}

func closeAll(mbConsumer, mbProducer messagebroker.Client, db *storage.DB, otherClosers ...func() error) func() error {
	return func() error {
		err1 := errors.Wrap(mbConsumer.Close(), "closing message broker consumer connection failed")
		err2 := errors.Wrap(db.Close(), "closing db connection failed")
		err3 := errors.Wrap(mbProducer.Close(), "closing message broker producer connection failed")
		errs := make([]error, 0, 1+1+1+len(otherClosers))
		errs = append(errs, err1, err2, err3)
		for _, closeOther := range otherClosers {
			if err := closeOther(); err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Wrap(multierror.Append(nil, errs...).ErrorOrNil(), "failed to close resources")
	}
}

func (p *processor) CheckHealth(ctx context.Context) error {
	if err := p.db.Ping(ctx); err != nil {
		return errors.Wrap(err, "[health-check] failed to ping DB")
	}
	type ts struct {
		TS *time.Time `json:"ts"`
	}
	now := ts{TS: time.Now()}
	bytes, err := json.MarshalContext(ctx, now)
	if err != nil {
		return errors.Wrapf(err, "[health-check] failed to marshal %#v", now)
	}
	responder := make(chan error, 1)
	p.mb.SendMessage(ctx, &messagebroker.Message{
		Headers: map[string]string{"producer": "santa"},
		Key:     p.cfg.MessageBroker.Topics[0].Name,
		Topic:   p.cfg.MessageBroker.Topics[0].Name,
		Value:   bytes,
	}, responder)

	return errors.Wrapf(<-responder, "[health-check] failed to send health check message to broker")
}

func runConcurrently[ARG any](ctx context.Context, run func(context.Context, ARG) error, args []ARG) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "unexpected deadline")
	}
	if len(args) == 0 {
		return nil
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(args))
	errChan := make(chan error, len(args))
	for i := range args {
		go func(ix int) {
			defer wg.Done()
			errChan <- errors.Wrapf(run(ctx, args[ix]), "failed to run:%#v", args[ix])
		}(i)
	}
	wg.Wait()
	close(errChan)
	errs := make([]error, 0, len(args))
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Wrap(multierror.Append(nil, errs...).ErrorOrNil(), "at least one execution failed")
}

func AreBadgesAchieved(actual *users.Enum[Type], expectedSubset ...Type) bool {
	if len(expectedSubset) == 0 {
		return actual == nil || len(*actual) == 0
	}
	if (actual == nil || len(*actual) == 0) && len(expectedSubset) > 0 {
		return false
	}
	for _, expectedType := range expectedSubset {
		var achieved bool
		for _, achievedType := range *actual {
			if achievedType == expectedType {
				achieved = true

				break
			}
		}
		if !achieved {
			return false
		}
	}

	return true
}

func IsBadgeGroupAchieved(actual *users.Enum[Type], expectedGroupType GroupType) bool {
	if actual == nil || len(*actual) == 0 {
		return false
	}
	for _, expectedType := range AllGroups[expectedGroupType] {
		var achieved bool
		for _, achievedType := range *actual {
			if achievedType == expectedType {
				achieved = true

				break
			}
		}
		if !achieved {
			return false
		}
	}

	return true
}

func requestingUserID(ctx context.Context) (requestingUserID string) {
	requestingUserID, _ = ctx.Value(requestingUserIDCtxValueKey).(string) //nolint:errcheck,revive // Not needed.

	return
}

func loadBadges(cfg *config) {
	LevelTypeOrder = make(map[Type]int, len(cfg.Levels))
	CoinTypeOrder = make(map[Type]int, len(cfg.Coins))
	SocialTypeOrder = make(map[Type]int, len(cfg.Socials))
	AllTypeOrder = make(map[Type]int, len(cfg.Levels)+len(cfg.Coins)+len(cfg.Socials))
	LevelTypeNames = make(map[Type]string, len(cfg.Levels))
	CoinTypeNames = make(map[Type]string, len(cfg.Coins))
	SocialTypeNames = make(map[Type]string, len(cfg.Socials))
	GroupTypeForEachType = make(map[Type]GroupType, len(cfg.Levels)+len(cfg.Coins)+len(cfg.Socials))
	AllNames = make(map[GroupType]map[Type]string, len(GroupsOrderSummaries))
	AllGroups = make(map[GroupType][]Type, len(GroupsOrderSummaries))
	Milestones = make(map[Type]AchievingRange, len(cfg.Levels)+len(cfg.Coins)+len(cfg.Socials))

	loadBadgesInfo(cfg.Levels, LevelGroupType)
	AllNames[LevelGroupType] = make(map[Type]string, len(LevelTypeNames))
	for key, val := range LevelTypeNames {
		AllNames[LevelGroupType][key] = val
	}
	loadBadgesInfo(cfg.Coins, CoinGroupType)
	AllNames[CoinGroupType] = make(map[Type]string, len(CoinTypeNames))
	for key, val := range CoinTypeNames {
		AllNames[CoinGroupType][key] = val
	}
	loadBadgesInfo(cfg.Socials, SocialGroupType)
	AllNames[SocialGroupType] = make(map[Type]string, len(SocialTypeNames))
	for key, val := range SocialTypeNames {
		AllNames[SocialGroupType][key] = val
	}
}

func loadBadgesInfo(badgeInfoList []*AchievingRange, groupType GroupType) {
	offset := len(AllTypes)
	for idx, val := range badgeInfoList {
		typeName := getTypeName(groupType)
		tpe := Type(fmt.Sprintf("%v%v", typeName, idx+1))
		AllTypes = append(AllTypes, tpe)
		Milestones[tpe] = *val

		switch groupType {
		case LevelGroupType:
			LevelTypeOrder[tpe] = idx
			LevelTypeNames[tpe] = val.Name
		case CoinGroupType:
			CoinTypeOrder[tpe] = idx
			CoinTypeNames[tpe] = val.Name
		case SocialGroupType:
			SocialTypeOrder[tpe] = idx
			SocialTypeNames[tpe] = val.Name
		default:
			log.Panic("wrong group type")
		}
		AllTypeOrder[tpe] = idx + offset
		GroupTypeForEachType[tpe] = groupType
		AllGroups[groupType] = append(AllGroups[groupType], tpe)
	}
}

func getTypeName(groupType GroupType) (typeName string) {
	switch groupType {
	case LevelGroupType:
		typeName = "l"
	case CoinGroupType:
		typeName = "c"
	case SocialGroupType:
		typeName = "s"
	default:
		log.Panic("wrong group type")
	}

	return
}
