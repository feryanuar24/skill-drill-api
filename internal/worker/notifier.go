package worker

import (
	"context"
	"log"
	"time"

	"skill-drill-api/internal/repository"
	"skill-drill-api/pkg/fcm"

	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
)

type Notifier struct {
	Repo     *repository.Repository
	FCM      *fcm.Client
	Location *time.Location
}

func (n *Notifier) Start() *cron.Cron {
	c := cron.New(cron.WithLocation(n.Location))
	_, _ = c.AddFunc("0,30 18-23 * * *", func() {
		if err := n.Notify(context.Background()); err != nil {
			log.Printf("notification worker: %v", err)
		}
	})
	c.Start()
	return c
}

func (n *Notifier) Notify(ctx context.Context) error {
	if n.FCM == nil {
		return nil
	}
	users, err := n.Repo.UsersToNotify(ctx, time.Now().In(n.Location))
	if err != nil {
		return err
	}
	group, ctx := errgroup.WithContext(ctx)
	for _, user := range users {
		user := user
		if user.FCMToken == nil {
			continue
		}
		group.Go(func() error {
			return n.FCM.Send(ctx, *user.FCMToken, "OmniPrep", "Your streak is about to be lost! Open OmniPrep now!")
		})
	}
	return group.Wait()
}
