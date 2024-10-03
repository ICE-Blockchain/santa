// SPDX-License-Identifier: ice License 1.0

package tasks

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/ice-blockchain/eskimo/users"
	appcfg "github.com/ice-blockchain/wintr/config"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/log"
	"github.com/ice-blockchain/wintr/time"
)

func New(ctx context.Context, _ context.CancelFunc) Repository {
	var cfg config
	appcfg.MustLoadFromKey(applicationYamlKey, &cfg)

	db := storage.MustConnect(ctx, ddl, applicationYamlKey)

	return &repository{
		cfg:      &cfg,
		shutdown: db.Close,
		db:       db,
	}
}

func StartProcessor(ctx context.Context, cancel context.CancelFunc) Processor {
	var cfg config
	appcfg.MustLoadFromKey(applicationYamlKey, &cfg)

	var mbConsumer messagebroker.Client
	prc := &processor{repository: &repository{
		cfg: &cfg,
		db:  storage.MustConnect(ctx, ddl, applicationYamlKey),
		mb:  messagebroker.MustConnect(ctx, applicationYamlKey),
	}}
	//nolint:contextcheck // It's intended. Cuz we want to close everything gracefully.
	mbConsumer = messagebroker.MustConnectAndStartConsuming(context.Background(), cancel, applicationYamlKey,
		&tryCompleteTasksCommandSource{processor: prc},
		&userTableSource{processor: prc},
		&miningSessionSource{processor: prc},
		&friendsInvitedSource{processor: prc},
	)
	prc.shutdown = closeAll(mbConsumer, prc.mb, prc.db)
	if cfg.TasksV2Enabled {
		prc.repository.loadTaskTranslationTemplates(cfg.TenantName)
	}

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
	val, err := json.MarshalContext(ctx, now)
	if err != nil {
		return errors.Wrapf(err, "[health-check] failed to marshal %#v", now)
	}
	responder := make(chan error, 1)
	p.mb.SendMessage(ctx, &messagebroker.Message{
		Headers: map[string]string{"producer": "santa"},
		Key:     p.cfg.MessageBroker.Topics[0].Name,
		Topic:   p.cfg.MessageBroker.Topics[0].Name,
		Value:   val,
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

func AreTasksCompleted(actual *users.Enum[Type], expectedSubset ...Type) bool {
	if len(expectedSubset) == 0 {
		return actual == nil || len(*actual) == 0
	}
	if (actual == nil || len(*actual) == 0) && len(expectedSubset) > 0 {
		return false
	}
	for _, expectedType := range expectedSubset {
		var completed bool
		for _, completedType := range *actual {
			if completedType == expectedType {
				completed = true

				break
			}
		}
		if !completed {
			return false
		}
	}

	return true
}

//nolint:funlen // .
func (r *repository) loadTaskTranslationTemplates(tenantName string) {
	const totalLanguages = 50
	allTaskTemplates = make(map[Type]map[languageCode]*taskTemplate, len(r.cfg.TasksList))
	for ix := range r.cfg.TasksList {
		var fileName string
		switch r.cfg.TasksList[ix].Group {
		case TaskGroupBadgeSocial, TaskGroupBadgeCoin, TaskGroupBadgeLevel, TaskGroupLevel, TaskGroupInviteFriends, TaskGroupMiningStreak:
			fileName = r.cfg.TasksList[ix].Group
		default:
			fileName = r.cfg.TasksList[ix].Type
		}
		content, fErr := translations.ReadFile(fmt.Sprintf("translations/%v/%v.json", strings.ToLower(tenantName), fileName))
		log.Panic(fErr) //nolint:revive // Wrong.
		allTaskTemplates[Type(r.cfg.TasksList[ix].Type)] = make(map[languageCode]*taskTemplate, totalLanguages)
		var languageData map[string]*struct {
			Title            string `json:"title"`
			ShortDescription string `json:"shortDescription"`
			LongDescription  string `json:"longDescription"`
			ErrorDescription string `json:"errorDescription"`
		}
		log.Panic(json.Unmarshal(content, &languageData))
		for language, data := range languageData {
			var tmpl taskTemplate
			tmpl.ShortDescription = data.ShortDescription
			tmpl.LongDescription = data.LongDescription
			tmpl.ErrorDescription = data.ErrorDescription
			tmpl.Title = data.Title
			tmpl.title = template.Must(template.New(fmt.Sprintf("task_%v_%v_title", r.cfg.TasksList[ix].Type, language)).Parse(data.Title))
			tmpl.shortDescription = template.Must(template.New(fmt.Sprintf("task_%v_%v_short_description", r.cfg.TasksList[ix].Type, language)).Parse(data.ShortDescription)) //nolint:lll // .
			tmpl.longDescription = template.Must(template.New(fmt.Sprintf("task_%v_%v_long_description", r.cfg.TasksList[ix].Type, language)).Parse(data.LongDescription))    //nolint:lll // .
			tmpl.errorDescription = template.Must(template.New(fmt.Sprintf("task_%v_%v_error_description", r.cfg.TasksList[ix].Type, language)).Parse(data.ErrorDescription)) //nolint:lll // .

			allTaskTemplates[Type(r.cfg.TasksList[ix].Type)][language] = &tmpl
		}
	}
}

func (t *taskTemplate) getTitle(data any) string {
	if data == nil {
		return t.Title
	}
	bf := new(bytes.Buffer)
	log.Panic(errors.Wrapf(t.title.Execute(bf, data), "failed to execute title template for data:%#v", data))

	return bf.String()
}

func (t *taskTemplate) getShortDescription(data any) string {
	if data == nil {
		return t.ShortDescription
	}
	bf := new(bytes.Buffer)
	log.Panic(errors.Wrapf(t.shortDescription.Execute(bf, data), "failed to execute short description template for data:%#v", data))

	return bf.String()
}

func (t *taskTemplate) getLongDescription(data any) string {
	if data == nil {
		return t.LongDescription
	}
	bf := new(bytes.Buffer)
	log.Panic(errors.Wrapf(t.longDescription.Execute(bf, data), "failed to execute long description template for data:%#v", data))

	return bf.String()
}

func (t *taskTemplate) getErrorDescription(data any) string {
	if data == nil {
		return t.ErrorDescription
	}
	bf := new(bytes.Buffer)
	log.Panic(errors.Wrapf(t.errorDescription.Execute(bf, data), "failed to execute error description template for data:%#v", data))

	return bf.String()
}
