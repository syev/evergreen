package evergreen

import (
	"fmt"

	"github.com/evergreen-ci/evergreen/db"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

// AuthUser configures a user for our Naive authentication setup.
type AuthUser struct {
	Username    string `bson:"username" json:"username" yaml:"username"`
	DisplayName string `bson:"display_name" json:"display_name" yaml:"display_name"`
	Password    string `bson:"password" json:"password" yaml:"password"`
	Email       string `bson:"email" json:"email" yaml:"email"`
}

// NaiveAuthConfig contains a list of AuthUsers from the settings file.
type NaiveAuthConfig struct {
	Users []*AuthUser `bson:"users" json:"users" yaml:"users"`
}

// CrowdConfig holds settings for interacting with Atlassian Crowd.
type CrowdConfig struct {
	Username string `bson:"username" json:"username" yaml:"username"`
	Password string `bson:"password" json:"password" yaml:"password"`
	Urlroot  string `bson:"url_root" json:"url_root" yaml:"urlroot"`
}

// LDAPConfig contains settings for interacting with an LDAP server.
type LDAPConfig struct {
	URL                string `bson:"url" json:"url" yaml:"url"`
	Port               string `bson:"port" json:"port" yaml:"port"`
	UserPath           string `bson:"path" json:"path" yaml:"path"`
	ServicePath        string `bson:"service_path" json:"service_path" yaml:"service_path"`
	Group              string `bson:"group" json:"group" yaml:"group"`
	ExpireAfterMinutes string `bson:"expire_after_minutes" json:"expire_after_minutes" yaml:"expire_after_minutes"`
}

// GithubAuthConfig holds settings for interacting with Github Authentication including the
// ClientID, ClientSecret and CallbackUri which are given when registering the application
// Furthermore,
type GithubAuthConfig struct {
	ClientId     string   `bson:"client_id" json:"client_id" yaml:"client_id"`
	ClientSecret string   `bson:"client_secret" json:"client_secret" yaml:"client_secret"`
	Users        []string `bson:"users" json:"users" yaml:"users"`
	Organization string   `bson:"organization" json:"organization" yaml:"organization"`
}

// AuthConfig has a pointer to either a CrowConfig or a NaiveAuthConfig.
type AuthConfig struct {
	LDAP   *LDAPConfig       `bson:"ldap,omitempty" json:"ldap" yaml:"ldap"`
	Crowd  *CrowdConfig      `bson:"crowd,omitempty" json:"crowd" yaml:"crowd"`
	Naive  *NaiveAuthConfig  `bson:"naive,omitempty" json:"naive" yaml:"naive"`
	Github *GithubAuthConfig `bson:"github,omitempty" json:"github" yaml:"github"`
}

func (c *AuthConfig) SectionId() string { return "auth" }

func (c *AuthConfig) Get() error {
	err := db.FindOneQ(ConfigCollection, db.Query(byId(c.SectionId())), c)
	if err != nil && err.Error() == errNotFound {
		*c = AuthConfig{}
		return nil
	}
	return errors.Wrapf(err, "error retrieving section %s", c.SectionId())
}

func (c *AuthConfig) Set() error {
	_, err := db.Upsert(ConfigCollection, byId(c.SectionId()), bson.M{
		"$set": bson.M{
			"crowd":  c.Crowd,
			"ldap":   c.LDAP,
			"naive":  c.Naive,
			"github": c.Github,
		},
	})
	return errors.Wrapf(err, "error updating section %s", c.SectionId())
}

func (c *AuthConfig) ValidateAndDefault() error {
	catcher := grip.NewSimpleCatcher()
	if c.Crowd == nil && c.LDAP == nil && c.Naive == nil && c.Github == nil {
		catcher.Add(errors.New("You must specify one form of authentication"))
	}
	if c.Naive != nil {
		used := map[string]bool{}
		for _, x := range c.Naive.Users {
			if used[x.Username] {
				catcher.Add(fmt.Errorf("Duplicate user %s in list", x.Username))
			}
			used[x.Username] = true
		}
	}
	if c.Github != nil {
		if c.Github.Users == nil && c.Github.Organization == "" {
			catcher.Add(errors.New("Must specify either a set of users or an organization for Github Authentication"))
		}
	}
	return catcher.Resolve()
}
