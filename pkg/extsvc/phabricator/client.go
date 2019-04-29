// Package phabricator is a package to interact with a Phabricator instance and its Conduit API.
package phabricator

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/uber/gonduit"
	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/requests"
)

// A Client provides high level methods to a Phabricator Conduit API.
type Client struct {
	conn *gonduit.Conn
}

// NewClient returns an authenticated Client, using the given URL and
// token. If provided, cli will be used to perform the underlying HTTP requests.
// This constructor needs a context because it calls the Conduit API to negotiate
// capabilities as part of the dial process.
func NewClient(ctx context.Context, url, token string, cli httpcli.Doer) (*Client, error) {
	conn, err := gonduit.DialContext(ctx, url, &core.ClientOptions{
		APIToken: token,
		Client:   httpcli.HeadersMiddleware("User-Agent", "sourcegraph/phabricator-client")(cli),
	})

	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

// Repo represents a single code repository.
type Repo struct {
	ID           uint64
	PHID         string
	Name         string
	VCS          string
	Callsign     string
	Shortname    string
	Status       string
	DateCreated  time.Time
	DateModified time.Time
	ViewPolicy   string
	EditPolicy   string
	URIs         []*URI
}

// URI of a Repository
type URI struct {
	ID   string
	PHID string

	Display    string
	Effective  string
	Normalized string

	Disabled bool

	BuiltinProtocol   string
	BuiltinIdentifier string

	DateCreated  time.Time
	DateModified time.Time
}

//
// Marshaling types
//

type apiRepo struct {
	ID          uint64             `json:"id"`
	PHID        string             `json:"phid"`
	Fields      apiRepoFields      `json:"fields"`
	Attachments apiRepoAttachments `json:"attachments"`
}

type apiRepoFields struct {
	Name         string        `json:"name"`
	VCS          string        `json:"vcs"`
	Callsign     string        `json:"callsign"`
	Shortname    string        `json:"shortname"`
	Status       string        `json:"status"`
	Policy       apiRepoPolicy `json:"policy"`
	DateCreated  unixTime      `json:"dateCreated"`
	DateModified unixTime      `json:"dateModified"`
}

type apiRepoPolicy struct {
	View string `json:"view"`
	Edit string `json:"edit"`
}

type apiRepoAttachments struct {
	URIs apiURIsContainer `json:"uris"`
}

type apiURIsContainer struct {
	URIs []apiURI `json:"uris"`
}

type apiURI struct {
	ID     string       `json:"id"`
	PHID   string       `json:"phid"`
	Fields apiURIFields `json:"fields"`
}

type apiURIFields struct {
	URI          apiURIs      `json:"uri"`
	Builtin      apiURIBultin `json:"builtin"`
	Disabled     bool         `json:"disabled"`
	DateCreated  unixTime     `json:"dateCreated"`
	DateModified unixTime     `json:"dateModified"`
}

type apiURIs struct {
	Display    string `json:"display"`
	Effective  string `json:"effective"`
	Normalized string `json:"normalized"`
}

type apiURIBultin struct {
	Protocol   string `json:"protocol"`
	Identifier string `json:"identifier"`
}

// Cursor represents the pagination cursor on many responses.
type Cursor struct {
	Limit  uint64 `json:"limit,omitempty"`
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
	Order  string `json:"order,omitempty"`
}

// ListReposArgs defines the constraints to be satisfied
// by the ListRepos method.
type ListReposArgs struct {
	*Cursor
}

// ListRepos lists all repositories matching the given arguments.
func (c *Client) ListRepos(ctx context.Context, args ListReposArgs) ([]*Repo, *Cursor, error) {
	var req struct {
		requests.Request
		ListReposArgs
		Attachments struct {
			URIs bool `json:"uris"`
		} `json:"attachments"`
	}

	req.ListReposArgs = args
	req.Attachments.URIs = true

	if req.Cursor == nil {
		req.Cursor = new(Cursor)
	}

	if req.Cursor.Order == "" {
		req.Cursor.Order = "oldest"
	}

	if req.Cursor.Limit == 0 {
		req.Cursor.Limit = 100
	}

	var res struct {
		Data   []apiRepo `json:"data"`
		Cursor Cursor    `json:"cursor"`
	}

	err := c.conn.CallContext(ctx, "diffusion.repository.search", &req, &res)
	if err != nil {
		return nil, nil, err
	}

	rs := make([]*Repo, 0, len(res.Data))
	for _, r := range res.Data {
		repo := &Repo{
			ID:           r.ID,
			PHID:         r.PHID,
			Name:         r.Fields.Name,
			VCS:          r.Fields.VCS,
			Callsign:     r.Fields.Callsign,
			Shortname:    r.Fields.Shortname,
			Status:       r.Fields.Status,
			ViewPolicy:   r.Fields.Policy.View,
			EditPolicy:   r.Fields.Policy.Edit,
			DateCreated:  time.Time(r.Fields.DateCreated),
			DateModified: time.Time(r.Fields.DateModified),
			URIs:         make([]*URI, 0, len(r.Attachments.URIs.URIs)),
		}

		for _, u := range r.Attachments.URIs.URIs {
			repo.URIs = append(repo.URIs, &URI{
				ID:                u.ID,
				PHID:              u.PHID,
				Display:           u.Fields.URI.Display,
				Effective:         u.Fields.URI.Effective,
				Normalized:        u.Fields.URI.Normalized,
				Disabled:          u.Fields.Disabled,
				BuiltinProtocol:   u.Fields.Builtin.Protocol,
				BuiltinIdentifier: u.Fields.Builtin.Identifier,
				DateCreated:       time.Time(u.Fields.DateCreated),
				DateModified:      time.Time(u.Fields.DateModified),
			})
		}

		rs = append(rs, repo)
	}

	return rs, &res.Cursor, nil
}

// GetRawDiff retrieves the raw diff of the diff with the given id.
func (c *Client) GetRawDiff(ctx context.Context, diffID int) (diff string, err error) {
	type request struct {
		requests.Request
		DiffID int `json:"diffID"`
	}

	req := request{DiffID: diffID}
	err = c.conn.CallContext(ctx, "differential.getrawdiff", &req, &diff)
	if err != nil {
		return "", err
	}

	return diff, nil
}

// DiffInfo contains information for a diff such as the author
type DiffInfo struct {
	Message     string    `json:"description"`
	AuthorName  string    `json:"authorName"`
	AuthorEmail string    `json:"authorEmail"`
	DateCreated string    `json:"dateCreated"`
	Date        time.Time `json:"omitempty"`
}

// GetDiffInfo retrieves the DiffInfo of the diff with the given id.
func (c *Client) GetDiffInfo(ctx context.Context, diffID int) (*DiffInfo, error) {
	type request struct {
		requests.Request
		IDs []int `json:"ids"`
	}

	req := request{IDs: []int{diffID}}

	var res map[string]*DiffInfo
	err := c.conn.CallContext(ctx, "differential.querydiffs", &req, &res)
	if err != nil {
		return nil, err
	}

	info, ok := res[strconv.Itoa(diffID)]
	if !ok {
		return nil, errors.Errorf("phabricator error: no diff info found for diff %d", diffID)
	}

	date, err := ParseDate(info.DateCreated)
	if err != nil {
		return nil, err
	}

	info.Date = date

	return info, nil
}

type unixTime time.Time

func (t *unixTime) UnmarshalJSON(data []byte) error {
	ts := string(data)

	// Ignore null, like in the main JSON package.
	if ts == "null" {
		return nil
	}

	d, err := ParseDate(strings.Trim(ts, `"`))
	if err != nil {
		return err
	}

	*t = unixTime(d)
	return nil
}

// ParseDate parses the given unix timestamp into a time.Time pointer.
func ParseDate(secStr string) (time.Time, error) {
	seconds, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "phabricator: could not parse date")
	}
	return time.Unix(seconds, 0).UTC(), nil
}
