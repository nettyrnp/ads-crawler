package entity

import (
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"time"
)

var (
	//reEmail      = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
	reComment  = regexp.MustCompile(`#[^#;]+(;|$)`)
	reProvider = regexp.MustCompile(`^ *([-_.\w]+), *([-_\w]+), *(direct|reseller)(, *([-_.\w]+))? *$`)
)

type PortalSortField int

const (
	SortByDomain PortalSortField = iota
	SortByCreationDate
)

var ErrInvalidLine = errors.New("invalid line")

type TPPKind string

const (
	KindCrawler TPPKind = "Crawler"
)

//func ValidateEmail(email string) error {
//	if email == "" {
//		return errors.New("empty email")
//	}
//	if !reEmail.MatchString(email) {
//		return errors.New("invalid email")
//	}
//	return nil
//}

type Portal struct {
	ID int `json:"-", db:"id"`
	//RawURL string       `json:"rawURL", db:"raw_url"`
	Protocol      string    `json:"protocol", db:"protocol"`
	CanonicalName string    `json:"canonicalName", db:"canonical_name"`
	Email         string    `json:"email", db:"email"`
	Phone         string    `json:"phone", db:"phone"`
	CertInfo      string    `json:"certInfo", db:"cert_info"`
	CreatedAt     time.Time `json:"createdAt",db:"created_at"`
}

func (c *Portal) Validate() error {
	var errs []error
	if len(c.CanonicalName) == 0 {
		errs = append(errs, errors.New("CanonicalName cannot be empty"))
	}
	if len(c.Protocol) == 0 {
		errs = append(errs, errors.New("Protocol cannot be empty"))
	}
	//if c.Provider != nil {
	//	errs = append(errs, c.Provider.validate()...)
	//}
	if len(errs) > 0 {
		return common.JoinErrors(errs)
	}
	return nil
}

type Provider struct {
	ID          int       `json:"-", db:"id"`
	DomainName  string    `json:"domainName", db:"domain_name"`
	AccountID   string    `json:"accountID", db:"account_id"`
	AccountType string    `json:"accountType", db:"account_Type"`
	CertAuthID  string    `json:"certAuthID", db:"cert_auth_id"`
	PortalID    int       `json:"portalID", db:"portal_id"`
	CreatedAt   time.Time `json:"createdAt", db:"created_at"`
}

func (c *Provider) validate() []error {
	var errs []error
	if len(c.DomainName) == 0 {
		errs = append(errs, errors.New("Provider name cannot be empty"))
	}
	if len(c.AccountID) == 0 {
		errs = append(errs, errors.New("AccountID cannot be empty"))
	}
	if len(c.AccountType) == 0 {
		errs = append(errs, errors.New("AccountType cannot be empty"))
	}
	//
	//// Email
	//if err := ValidateEmail(c.Email); err != nil {
	//	errs = append(errs, err)
	//}
	//
	//// Password
	//if len(c.Password) < 8 {
	//	errs = append(errs, errors.New("Password should be at least 8 symbols long"))
	//}
	//if !reOneCapital.MatchString(c.Password) || !reOneSmall.MatchString(c.Password) || !reOneDigit.MatchString(c.Password) || !reOneSpecial.MatchString(c.Password) {
	//	errs = append(errs, errors.New("Password should contain only latin characters and contain at least 1 letter in upper case and at least 1 digit"))
	//}
	//
	//if len(c.Phone) == 0 {
	//	errs = append(errs, errors.New("User phone cannot be empty"))
	//}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ParseProvider(s string) (*Provider, error) {
	actual := reComment.ReplaceAllString(s, "")
	actual = strings.ToLower(actual)

	arr := reProvider.FindStringSubmatch(actual)
	if len(arr) == 6 {
		e := &Provider{
			DomainName:  reProvider.FindStringSubmatch(actual)[1],
			AccountID:   reProvider.FindStringSubmatch(actual)[2],
			AccountType: reProvider.FindStringSubmatch(actual)[3],
			CertAuthID:  reProvider.FindStringSubmatch(actual)[5],
		}
		//todo: e.validate()
		return e, nil
	}
	if len(arr) == 4 {
		e := &Provider{
			DomainName:  reProvider.FindStringSubmatch(actual)[1],
			AccountID:   reProvider.FindStringSubmatch(actual)[2],
			AccountType: reProvider.FindStringSubmatch(actual)[3],
		}
		//todo: e.validate()
		return e, nil
	}
	return nil, ErrInvalidLine
}
