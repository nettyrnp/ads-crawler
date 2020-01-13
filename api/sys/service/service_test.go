package service

import (
	"context"
	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"github.com/nettyrnp/ads-crawler/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nettyrnp/ads-crawler/api/sys/repository"
)

type smsTestNotifier struct {
	code string
}

func (n *smsTestNotifier) Send(ctx context.Context, toAddr, code string) error {
	n.code = code
	return nil
}

type emailTestNotifier struct {
	code  string
	email string
}

func (n *emailTestNotifier) Send(ctx context.Context, email string, msg string) error {
	n.code = msg
	n.email = email
	return nil
}

func TestService(t *testing.T) {
	t.Parallel()

	repo, closer, repoErr := repository.NewDockerRepo()
	defer closer()
	require.NoError(t, repoErr)
	require.NotNil(t, repo)

	t.Run("get portals", testGetPortals(repo))
	t.Run("get providers by portal", testGetProvidersByPortal(repo))
}

func testGetPortals(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := New(config.Config{}, "", repo, &smsTestNotifier{}, &emailTestNotifier{})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		ps, err := svc.GetPortals(ctx)
		require.NoError(t, err)
		assert.Len(t, ps, 5)
	}
}

func testGetProvidersByPortal(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := New(config.Config{}, "", repo, &smsTestNotifier{}, &emailTestNotifier{})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		testProvName1 := "cnn.com"
		prs, err := svc.GetProvidersByPortal(ctx, testProvName1)
		require.NoError(t, err)
		assert.Len(t, prs, 0)

		testProvName2 := "nytimes.com"
		prs, err = svc.GetProvidersByPortal(ctx, testProvName2)
		require.NoError(t, err)
		assert.Len(t, prs, 0)

		testPortIdCNN := 1
		testPortIdNYTimes := 3

		testProv1 := entity.Provider{
			ID:          123,
			DomainName:  testProvName1,
			AccountID:   "acc1",
			AccountType: "direct",
			CertAuthID:  "cert1",
			PortalID:    testPortIdCNN,
			CreatedAt:   time.Now().UTC(),
		}
		id, err := svc.AddProvider(ctx, &testProv1)
		require.NoError(t, err)
		assert.Equal(t, 1, id)

		testProv21 := entity.Provider{
			ID:          4561,
			DomainName:  testProvName2,
			AccountID:   "acc2",
			AccountType: "reseller",
			CertAuthID:  "cert2",
			PortalID:    testPortIdNYTimes,
			CreatedAt:   time.Now().UTC(),
		}
		id, err = svc.AddProvider(ctx, &testProv21)
		require.NoError(t, err)
		assert.Equal(t, 2, id)

		testProv22 := entity.Provider{
			ID:          4562,
			DomainName:  testProvName2,
			AccountID:   "acc2",
			AccountType: "direct",
			CertAuthID:  "cert2",
			PortalID:    testPortIdNYTimes,
			CreatedAt:   time.Now().UTC(),
		}
		id, err = svc.AddProvider(ctx, &testProv22)
		require.NoError(t, err)
		assert.Equal(t, 3, id)

		prs, err = svc.GetProvidersByPortal(ctx, testProvName1)
		require.NoError(t, err)
		assert.Len(t, prs, 1)

		prs, err = svc.GetProvidersByPortal(ctx, testProvName2)
		require.NoError(t, err)
		assert.Len(t, prs, 2)

	}
}
