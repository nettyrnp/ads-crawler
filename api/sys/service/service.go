package service

import (
	"context"
	"fmt"
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"github.com/nettyrnp/ads-crawler/api/sys/repository"
	"github.com/nettyrnp/ads-crawler/config"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrUserNotConfirmed = errors.New("user not confirmed")
	ErrEmptyCustomerID  = errors.New("empty customer id")
)

type Service interface {
	GetPortals(ctx context.Context) ([]*entity.Portal, error)
	GetPortalsExt(ctx context.Context, from time.Time, till time.Time, sortBy entity.PortalSortField, sortDesc bool, limit uint64, offset uint64) ([]*entity.Portal, int, error)
	AddProvider(ctx context.Context, provider *entity.Provider) (int, error)
	DeleteProvider(ctx context.Context, portalID string) error
	GetProvidersByPortal(ctx context.Context, portalName string) ([]*entity.Provider, error)
	NotifyPortalAdmins(ctx context.Context, portal *entity.Portal) error
}

type emailNotifier interface {
	Send(ctx context.Context, toAddr, msg string) error
}

type smsNotifier interface {
	Send(ctx context.Context, toAddr, msg string) error
}

type AdsService struct {
	Name          string
	Repo          repository.Repository
	SmsNotifier   smsNotifier
	EmailNotifier emailNotifier
	Conf          config.Config
}

func New(conf config.Config, name string, r repository.Repository, sms smsNotifier, email emailNotifier) *AdsService {
	return &AdsService{
		Name:          name,
		Repo:          r,
		SmsNotifier:   sms,
		EmailNotifier: email,
		Conf:          conf,
	}
}

func (s *AdsService) GetPortals(ctx context.Context) ([]*entity.Portal, error) {
	return s.Repo.GetPortals(ctx)
}

func (s *AdsService) GetPortalsExt(ctx context.Context, from time.Time, till time.Time, sortBy entity.PortalSortField, sortDesc bool, limit uint64, offset uint64) ([]*entity.Portal, int, error) {
	opts := repository.PortalsQueryOpts{
		Limit:  limit,
		Offset: offset,
		Desc:   sortDesc,
		SortBy: sortBy,
		From:   from,
		To:     till,
	}
	return s.Repo.GetPortalsExt(ctx, opts)
}

func (s *AdsService) AddProvider(ctx context.Context, provider *entity.Provider) (int, error) {
	return s.Repo.AddProvider(ctx, provider)
}

func (s *AdsService) DeleteProvider(ctx context.Context, portalID string) error {
	return s.Repo.DeleteProvider(ctx, portalID)
}

func (s *AdsService) GetProvidersByPortal(ctx context.Context, portalName string) ([]*entity.Provider, error) {
	return s.Repo.GetProvidersByPortal(ctx, portalName)
}

func (s *AdsService) NotifyPortalAdmins(ctx context.Context, portal *entity.Portal) error {
	msg := fmt.Sprintf("Dear admins of poratl '%v', please be informed that your portal has no publicly available 'ads.txt' file!", portal.CanonicalName)
	var errs []error
	if portal.Email != "" {
		err := s.EmailNotifier.Send(ctx, portal.Email, msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if portal.Phone != "" {
		err := s.SmsNotifier.Send(ctx, portal.Phone, msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return common.JoinErrors(errs)
	}
	return nil
}
