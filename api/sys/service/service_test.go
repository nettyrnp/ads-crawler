package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"github.com/nettyrnp/ads-crawler/api/sys/repository"
	"github.com/nettyrnp/ads-crawler/api/sys/service/token"
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

func bodyCreator(code string) string {
	return code
}

func TestService(t *testing.T) {
	t.Parallel()

	repo, closer, repoErr := repository.NewDockerRepo()
	defer closer()
	require.NoError(t, repoErr)
	require.NotNil(t, repo)

	t.Run("sign-up", testServiceSignUp(repo))
	t.Run("tokens", testTokenService(repo))
	t.Run("get-consent by email", testGetConsentByEmail(repo))
	t.Run("add-consent invalid params", testAddConsentInvalidParams(repo))
	t.Run("multiple consents get", testMultipleConsentsGet(repo))
	t.Run("get consents invalid order-attribute", testMultipleConsentsInvalidOrderBy(repo))
	t.Run("consent by invalid OTP", testConsentByInvalidOTP(repo))
	t.Run("confirm flow", testConfirmFlow(repo))
	t.Run("confirm invalid params", testConfirmInvalidParams(repo))
}

func testServiceSignUp(repo repository.Repository) func(t *testing.T) {
	return func(t *testing.T) {
		smsNotifier := &smsTestNotifier{}
		svc := &AdsService{
			Name:        "psu",
			Repo:        repo,
			SmsNotifier: smsNotifier,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		e := entity.Portal{
			Password: "pwd",
			Email:    "test@test.com",
		}

		validTill := time.Now().Add(time.Minute)
		acc, signupErr := svc.StartSignup(ctx, e, "", validTill)
		require.NoError(t, signupErr)

		assert.Equal(t, "test@test.com", acc.Email)
		assert.Equal(t, entity.StatusUnconfirmed.String(), acc.Status)

		unconfirmed, signInErr := svc.Signin(ctx, *acc)
		assert.EqualError(t, signInErr, "user not confirmed")
		assert.Equal(t, entity.StatusUnconfirmed.String(), unconfirmed.Status)

		confirmedAcc, err := svc.FinishSignup(ctx, smsNotifier.code)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusConfirmed.String(), confirmedAcc.Status)
	}
}

func testTokenService(r repository.Repository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := token.New(r)

		tokens := map[string]string{
			"james@metallica.com":   "metallica is cool",
			"carry.king@slayer.com": "slayer rocks",
			"dave@megadeth.com":     "who is laughing now James?",
			"ronnie@dio.com":        "Holy diver!!! w",
		}

		doConcurrently := func(tokens map[string]string, doStuff func(ctx context.Context, email string, token string) error) {
			var wg sync.WaitGroup
			wg.Add(len(tokens))
			rand.Seed(time.Now().UnixNano())
			const maxRetries = 5
			for k, v := range tokens {
				go func(email string, token string) {
					defer wg.Done()
					var retries int
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					for retries < maxRetries {
						err := doStuff(ctx, email, token)
						if err == nil {
							return
						}

						retries++
						if retries == maxRetries {
							t.Error(fmt.Errorf("max number of retries has been reached, last error: %s", err.Error()))
							return
						}

						milliSecs := rand.Int63n(100)
						t.Logf("%d attempt of retry to process token for %s, will retry in %d milliseconds", retries, email, milliSecs)
						time.Sleep(time.Duration(milliSecs) * time.Millisecond)
					}
				}(k, v)
			}
			wg.Wait()
		}

		doConcurrently(tokens, func(ctx context.Context, email string, token string) error {
			return svc.Save(ctx, email, token, time.Now().UTC().Add(time.Minute))
		})

		t.Log("check..")
		doConcurrently(tokens, func(ctx context.Context, email string, token string) error {
			tokenFromDB, err := svc.Check(ctx, email)
			if err != nil {
				return err
			}

			assert.Equal(t, token, tokenFromDB)
			return nil
		})

		t.Log("delete")
		doConcurrently(tokens, func(ctx context.Context, email string, token string) error {
			err := svc.Delete(ctx, email)
			assert.NoError(t, err)
			return err
		})

		t.Log("read removed")
		doConcurrently(tokens, func(ctx context.Context, email string, token string) error {
			tokenFromDB, err := svc.Check(ctx, email)
			assert.EqualError(t, err, sql.ErrNoRows.Error())
			assert.Empty(t, tokenFromDB)
			return nil
		})
	}
}

func testGetConsentByEmail(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		en := &emailTestNotifier{}
		svc := NewConsent(repo, en, bodyCreator)
		const customer = "jerry.smith"
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		for i := 0; i < 6; i++ {
			email := fmt.Sprintf("peter-%d@family.guy", rand.Int())
			require.NoError(t, svc.AddConsent(ctx, customer, email))
			require.Equal(t, email, en.email)
			byOTP, err := svc.ConsentByOTP(ctx, en.code)
			require.NoError(t, err)
			require.Equal(t, email, byOTP.PsuEmail)
		}
		for i := 0; i < 4; i++ {
			email := fmt.Sprintf("rick-%d@rick.and.morty", rand.Int())
			require.NoError(t, svc.AddConsent(ctx, customer, email))
			require.Equal(t, email, en.email)
			byOTP, err := svc.ConsentByOTP(ctx, en.code)
			require.NoError(t, err)
			require.Equal(t, email, byOTP.PsuEmail)
		}

		validate := func(t *testing.T, emailPattern string, c *entity.Consent) {
			assert.Equal(t, entity.ConsentPending, c.Status)
			assert.Contains(t, c.PsuEmail, emailPattern)
			assert.Equal(t, customer, c.CustomerID)
		}

		familyGuy, err := svc.ConsentsByEmail(ctx, customer, "family.guy")
		require.NoError(t, err)
		require.Len(t, familyGuy, 5)
		for _, cons := range familyGuy {
			validate(t, "family.guy", cons)
		}

		rickAndMorty, err := svc.ConsentsByEmail(ctx, customer, "rick.and.morty")
		require.NoError(t, err)
		require.Len(t, rickAndMorty, 4)
		for _, cons := range rickAndMorty {
			validate(t, "rick.and.morty", cons)
		}
	}
}

func testAddConsentInvalidParams(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := NewConsent(repo, &emailTestNotifier{}, bodyCreator)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert.EqualError(t, svc.AddConsent(ctx, "", ""), ErrEmptyCustomerID.Error())
		assert.EqualError(t, svc.AddConsent(ctx, "test-customer", ""), "empty email")
		assert.EqualError(t, svc.AddConsent(ctx, "test-customer", "invalid.email"), "invalid email")
	}
}

func testFilterByStatusWithLimits(svc *Consent, customerID string) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		allConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{},
			entity.ConsentPending.String(), entity.SortByStatus, false, 10, 0)
		require.NoError(t, err)
		require.Len(t, allConsents, 10)
		require.Equal(t, 24, n)
		for _, cons := range allConsents {
			assert.Equal(t, customerID, cons.CustomerID)
		}

		moreConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{},
			entity.ConsentPending.String(), entity.SortByCreatedDate, false, 10, 10)
		require.NoError(t, err)
		require.Len(t, moreConsents, 10)
		require.Equal(t, 24, n)
		for i, cons := range moreConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i+11), cons.PsuEmail)
			assert.Equal(t, entity.ConsentPending, cons.Status)
			assert.Equal(t, customerID, cons.CustomerID)
		}

		lastConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{},
			entity.ConsentPending.String(), entity.SortByCreatedDate, false, 5, 20)
		require.NoError(t, err)
		require.Len(t, lastConsents, 4)
		for i, cons := range lastConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i+21), cons.PsuEmail)
			assert.Equal(t, entity.ConsentPending, cons.Status)
			assert.Equal(t, customerID, cons.CustomerID)
		}

		confirmedConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{},
			entity.ConsentConfirmed.String(), entity.SortByCreatedDate, false, 5, 0)
		require.NoError(t, err)
		require.Len(t, confirmedConsents, 1)
		assert.Equal(t, "0@test.com", confirmedConsents[0].PsuEmail)
		assert.Equal(t, entity.ConsentConfirmed, confirmedConsents[0].Status)
		assert.Equal(t, customerID, confirmedConsents[0].CustomerID)
	}
}

func testFilterWithNoLimits(svc *Consent, customerID string) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		allConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{}, "", entity.SortByCreatedDate, false, 0, 0)
		require.NoError(t, err)
		require.Len(t, allConsents, 20)
		require.Equal(t, 25, n)
		for i, cons := range allConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i), cons.PsuEmail)
			assert.Equal(t, customerID, cons.CustomerID)
		}

		pendingConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{}, entity.ConsentPending.String(), entity.SortByCreatedDate, false, 0, 0)
		require.NoError(t, err)
		require.Len(t, pendingConsents, 20)
		require.Equal(t, 24, n)
		for i, cons := range pendingConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i+1), cons.PsuEmail)
			assert.Equal(t, entity.ConsentPending, cons.Status)
			assert.Equal(t, customerID, cons.CustomerID)
		}

		confirmedConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{}, entity.ConsentConfirmed.String(), entity.SortByCreatedDate, false, 0, 0)
		require.NoError(t, err)
		require.Len(t, confirmedConsents, 1)
		require.Equal(t, 1, n)
		assert.Equal(t, "0@test.com", confirmedConsents[0].PsuEmail)
		assert.Equal(t, entity.ConsentConfirmed, confirmedConsents[0].Status)
		assert.Equal(t, customerID, confirmedConsents[0].CustomerID)
	}
}

func testFilterByPeriodWithNoLimits(svc *Consent, customerID string, from time.Time, till time.Time) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		consents, n, err := svc.ConsentsByFilter(ctx, customerID, from, till, "", entity.SortByCreatedDate, false, 0, 0)
		require.NoError(t, err)
		require.Len(t, consents, 20)
		require.Equal(t, 25, n)
		for i, cons := range consents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i), cons.PsuEmail)
			if cons.PsuEmail == "0@test.com" {
				assert.Equal(t, entity.ConsentConfirmed, cons.Status)
			} else {
				assert.Equal(t, entity.ConsentPending, cons.Status)
			}
			assert.Equal(t, customerID, cons.CustomerID)
		}

		noConsents, n, err := svc.ConsentsByFilter(ctx, customerID, time.Now().UTC(), till, "", entity.SortByCreatedDate, false, 0, 0)
		assert.NoError(t, err)
		assert.Empty(t, noConsents)
		require.Zero(t, n)
	}
}

func testFilterByPeriodWithLimits(svc *Consent, customerID string, from time.Time, till time.Time) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		consents, n, err := svc.ConsentsByFilter(ctx, customerID, from, till, "", entity.SortByCreatedDate, false, 10, 0)
		require.NoError(t, err)
		require.Len(t, consents, 10)
		require.Equal(t, 25, n)
		for i, cons := range consents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i), cons.PsuEmail)
			if cons.PsuEmail == "0@test.com" {
				assert.Equal(t, entity.ConsentConfirmed, cons.Status)
			} else {
				assert.Equal(t, entity.ConsentPending, cons.Status)
			}
			assert.Equal(t, customerID, cons.CustomerID)
		}

		moreConsents, n, err := svc.ConsentsByFilter(ctx, customerID, from, till, "", entity.SortByCreatedDate, false, 10, 10)
		assert.NoError(t, err)
		require.Len(t, moreConsents, 10)
		require.Equal(t, 25, n)
		for i, cons := range moreConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i+10), cons.PsuEmail)
			assert.Equal(t, entity.ConsentPending, cons.Status)
			assert.Equal(t, customerID, cons.CustomerID)
		}

		lastConsents, n, err := svc.ConsentsByFilter(ctx, customerID, from, till, "", entity.SortByCreatedDate, false, 10, 20)
		assert.NoError(t, err)
		require.Len(t, lastConsents, 5)
		require.Equal(t, 25, n)
		for i, cons := range lastConsents {
			assert.Equal(t, fmt.Sprintf("%d@test.com", i+20), cons.PsuEmail)
			assert.Equal(t, entity.ConsentPending, cons.Status)
			assert.Equal(t, customerID, cons.CustomerID)
		}
	}
}

func testMultipleConsentsGet(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := NewConsent(repo, &emailTestNotifier{}, bodyCreator)
		createdTime := time.Now().UTC()
		customerID := uuid.NewV4().String()
		const consentsNum = 25
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for i := 0; i < consentsNum; i++ {
			uniqueEmail := fmt.Sprintf("%d@test.com", i)
			require.NoError(t, svc.AddConsent(ctx, customerID, uniqueEmail))
		}
		require.NoError(t, svc.Confirm(ctx, customerID, "0@test.com", "testbank", "test token"))

		t.Run("find with no limits", testFilterWithNoLimits(svc, customerID))
		t.Run("find all consents with limit", testFilterByStatusWithLimits(svc, customerID))

		t.Run("find by period with no limit", testFilterByPeriodWithNoLimits(svc, customerID, createdTime, createdTime.Add(time.Hour)))
		t.Run("find by period with limit", testFilterByPeriodWithLimits(svc, customerID, createdTime, createdTime.Add(time.Hour)))

		t.Run("wrong time boundaries", func(t *testing.T) {
			_, _, err := svc.ConsentsByFilter(ctx, customerID, createdTime.Add(time.Hour), time.Now().UTC(), "", entity.SortByStatus, false, 0, 0)
			assert.EqualError(t, err, ErrIncorrectTimeBoundaries.Error())
		})

		t.Run("no rejected consents", func(t *testing.T) {
			consents, _, err := svc.ConsentsByFilter(ctx, customerID, time.Time{}, time.Time{}, entity.ConsentRejected.String(), entity.SortByStatus, false, 0, 0)
			assert.NoError(t, err)
			assert.Empty(t, consents)
		})
	}
}

func testMultipleConsentsInvalidOrderBy(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := NewConsent(repo, &emailTestNotifier{}, bodyCreator)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, n, err := svc.ConsentsByFilter(ctx, "", time.Time{}, time.Time{}, "", -1, false, 0, 0)
		assert.Nil(t, res)
		assert.Zero(t, n)
		assert.EqualError(t, err, ErrWrongSortAttribute.Error())
	}
}

func testConsentByInvalidOTP(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := NewConsent(repo, &emailTestNotifier{}, bodyCreator)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := svc.ConsentByOTP(ctx, "")
		assert.EqualError(t, err, ErrEmptyOTP.Error())

		_, err = svc.ConsentByOTP(ctx, "this is invalid otp")
		assert.EqualError(t, err, sql.ErrNoRows.Error())
	}
}

func testConfirmFlow(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		en := &emailTestNotifier{
			code: "123",
		}
		svc := NewConsent(repo, en, bodyCreator)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		customerID := uuid.NewV4().String()
		email := "timmy@south.park"
		require.NoError(t, svc.AddConsent(ctx, customerID, email))

		consent, err := svc.ConsentByOTP(ctx, en.code)
		require.NoError(t, err)
		assert.Equal(t, customerID, consent.CustomerID)
		assert.Equal(t, email, consent.PsuEmail)
		assert.Equal(t, entity.ConsentPending, consent.Status)

		require.NoError(t, svc.Confirm(ctx, consent.CustomerID, consent.PsuEmail, "bank of America", "token of time"))

		consArr, err := svc.ConsentsByEmail(ctx, customerID, email)
		require.NoError(t, err)
		require.Len(t, consArr, 1)
		assert.Equal(t, entity.ConsentConfirmed, consArr[0].Status)
		assert.Equal(t, consent.CustomerID, consArr[0].CustomerID)
		assert.Equal(t, consent.PsuEmail, consArr[0].PsuEmail)
		assert.Equal(t, 24*90, int(consArr[0].Lifetime.Hours()))
	}
}

func testConfirmInvalidParams(repo *repository.RDBMSRepository) func(t *testing.T) {
	return func(t *testing.T) {
		svc := NewConsent(repo, &emailTestNotifier{}, bodyCreator)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert.EqualError(t, svc.Confirm(ctx, "", "", "", ""), ErrEmptyCustomerID.Error())
		assert.EqualError(t, svc.Confirm(ctx, "test-customer", "", "", ""), "empty email")
		assert.EqualError(t, svc.Confirm(ctx, "test-customer", "invalid.email", "", ""), "invalid email")
		assert.EqualError(t, svc.Confirm(ctx, "test-customer", "asd@asd.com", "", ""), ErrEmptyBank.Error())
		assert.EqualError(t, svc.Confirm(ctx, "test-customer", "asd@asd.com", "test-bank", ""), ErrEmptyToken.Error())
	}
}
