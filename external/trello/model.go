package trello

import (
	"encoding/json"
	"fmt"
	"time"
)

// Types

type DateTime time.Time

func (d *DateTime) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to unmarshal date to string")
	}
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		return fmt.Errorf("failed to parse date: %s", s)
	}
	*d = DateTime(t)
	return nil
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	formatted := t.Format("2006-01-02T15:04:05.000Z")
	return json.Marshal(formatted)
}

// Enums

type BoardPermissionLevel string

const (
	OrgBoardPermission   BoardPermissionLevel = `org`
	BoardBoardPermission BoardPermissionLevel = `board`
)

type CardAging string

const (
	RegularCardAging CardAging = `regular`
	PirateCardAging  CardAging = `pirate`
)

type BoardLimitAttachmentStatus string

const (
	BoardLimitOK      BoardLimitAttachmentStatus = `ok`
	BoardLimitWarning BoardLimitAttachmentStatus = `warning`
)

// Structs

// Board

type BoardResponse struct {
	ID                 string                   `json:"id"`
	Name               string                   `json:"name"`
	Description        string                   `json:"desc"`
	DescriptionData    string                   `json:"descData"`
	Closed             bool                     `json:"closed"`
	Creator            string                   `json:"idMemberCreator"`
	Organization       string                   `json:"idOrganization"`
	Pinned             bool                     `json:"pinned"`
	URL                string                   `json:"url"`
	ShortURL           string                   `json:"shortUrl"`
	Preferences        BoardPreferencesResponse `json:"prefs"`
	LabelNames         map[string]string        `json:"labelNames"`
	Limits             BoardLimitsResponse      `json:"limits"`
	Starred            bool                     `json:"starred"`
	Memberships        string                   `json:"memberships"`
	ShortLink          string                   `json:"shortLink"`
	Subscribed         bool                     `json:"subscribed"`
	PowerUps           string                   `json:"powerUps"`
	LastActivity       DateTime                 `json:"dateLastActivity"`
	LastViewed         DateTime                 `json:"dateLastView"`
	IDTags             string                   `json:"idTags"`
	DatePluginDisabled DateTime                 `json:"datePluginDisable"`
	CreationMethod     string                   `json:"creationMethod"`
	IXUpdate           int                      `json:"ixUpdate"`
	TemplateGallery    string                   `json:"templateGallery"`
	EnterpriseOwned    bool                     `json:"enterpriseOwned"`
}

type BoardPreferencesResponse struct {
	PermissionLevel        BoardPermissionLevel `json:"permissionLevel"`
	HideVotes              bool                 `json:"hideVotes"`
	Voting                 bool                 `json:"voting"`
	Comments               string               `json:"comments"`
	Invitations            any                  `json:"invitations"`
	SelfJoin               bool                 `json:"selfJoin"`
	CardCovers             bool                 `json:"cardCovers"`
	IsTemplate             bool                 `json:"isTemplate"`
	CardAging              CardAging            `json:"cardAging"`
	CalendarFeedEnabled    bool                 `json:"calendarFeedEnabled"`
	Background             string               `json:"background"`
	BackgroundImage        string               `json:"backgroundImage"`
	BackgroundImageScaling []struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		URL    string `json:"url"`
	} `json:"backgroundImageScaled"`
	BackgroundTile        bool   `json:"backgroundTile"`
	BackgroundBrightness  string `json:"backgroundBrightness"`
	BackgroundBottomColor string `json:"backgroundBottomColor"`
	BackgroundTopColor    string `json:"backgroundTopColor"`
	CanBePublic           bool   `json:"canBePublic"`
	CanBeEnterprise       bool   `json:"canBeEnterprise"`
	CanBeOrg              bool   `json:"canBeOrg"`
	CanBePrivate          bool   `json:"canBePrivate"`
	CanInvite             bool   `json:"canInvite"`
}

type BoardLimitsResponse struct {
	Attachments BoardLimitsAttachments `json:"attachments"`
}

type BoardLimitsAttachments struct {
	PerBoard struct {
		Status    BoardLimitAttachmentStatus `json:"status"`
		DisableAt int                        `json:"disableAt"`
		WarnAt    int                        `json:"warnAt"`
	} `json:"perBoard"`
}

type NewBoardRequest struct {
	Name                 string  `json:"name" validate:"min=1,max=16384"`
	DefaultLabels        *bool   `json:"defaultLabels,omitempty"`
	DefaultLists         *bool   `json:"defaultLists,omitempty"`
	Description          *string `json:"desc,omitempty" validate:"min=0,max=16384"`
	OrganizationID       *string `json:"idOrganization,omitempty" validate:"trelloID"`
	SourceBoardID        *string `json:"idBoardSource,omitempty" validate:"trelloID"`
	KeepFromSource       *string `json:"keepFromSource,omitempty" validate:"oneof=cards none"`
	PowerUps             *string `json:"powerUps,omitempty" validate:"oneof=all calendar cardAging recap voting"`
	PrefsPermissionLevel *string `json:"prefs_permissionLevel,omitempty" validate:"oneof=org private public"`
	PrefsVoting          *string `json:"prefs_voting,omitempty" validate:"oneof=disabled members observers org public"`
	PrefsComments        *string `json:"prefs_comments,omitempty" validate:"oneof=disabled members observers org public"`
	PrefsInvitations     *string `json:"prefs_invitations,omitempty" validate:"oneof=members admins"`
	PrefsSelfJoin        *bool   `json:"prefs_selfJoin,omitempty"`
	PrefsCardCovers      *bool   `json:"prefs_cardCovers,omitempty"`
	PrefsBackground      *string `json:"prefs_background,omitempty" validate:"oneof=blue orange green red purple pink lime sky grey"`
	PrefsCardAging       *string `json:"prefs_cardAging,omitempty" validate:"oneof=pirate regular"`
}

// Board Lists

type BoardListsResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Closed     bool                `json:"closed"`
	Position   int                 `json:"pos"`
	SoftLimit  string              `json:"softLimit"`
	BoardID    string              `json:"idBoard"`
	Subscribed bool                `json:"subscribed"`
	Limits     BoardLimitsResponse `json:"limits"`
}

// Card

type CardResponse struct {
	ID                     string                       `json:"id"`
	Badges                 CardBadgesResponse           `json:"badges"`
	CheckItemStates        []CardCheckItemStateResponse `json:"checkItemStates"`
	Closed                 bool                         `json:"closed"`
	LastActivity           DateTime                     `json:"dateLastActivity"`
	Description            string                       `json:"desc"`
	DescriptionData        CardDescDataResponse         `json:"descData"`
	Due                    *DateTime                    `json:"due"`
	DueComplete            bool                         `json:"dueComplete"`
	CoverImageAttachmentID *string                      `json:"idAttachmentCover"`
	BoardID                string                       `json:"idBoard"`
	ChecklistIDs           []string                     `json:"idChecklists"`
	LabelIDs               []string                     `json:"idLabels"`
	ListID                 string                       `json:"idList"`
	MemberIDs              []string                     `json:"idMembers"`
	VotedMemberIDs         []string                     `json:"idMembersVoted"`
	ShortID                int                          `json:"idShort"`
	Labels                 []CardLabelResponse          `json:"labels"`
	ManualCoverAttachment  bool                         `json:"manualCoverAttachment"`
	Name                   string                       `json:"name"`
	ListPosition           int                          `json:"pos"`
	ShortLink              string                       `json:"shortLink"`
	ShortURL               string                       `json:"shortUrl"`
	Start                  *DateTime                    `json:"start"`
	Subscribed             bool                         `json:"subscribed"`
	URL                    string                       `json:"url"`
	Address                string                       `json:"address"`
	LocationName           string                       `json:"locationName"`
	Coordinates            any                          `json:"coordinates"`
}

type CardDescDataResponse struct {
	Emoji any `json:"emoji"`
}

type CardBadgesResponse struct {
	Attachments           int                        `json:"attachments"`
	Fogbugz               string                     `json:"fogbugz"`
	CheckItems            int                        `json:"checkItems"`
	CheckItemsChecked     int                        `json:"checkItemsChecked"`
	CheckItemsEarliestDue *DateTime                  `json:"checkItemsEarliestDue"`
	Comments              int                        `json:"comments"`
	Description           bool                       `json:"description"`
	Due                   *DateTime                  `json:"due"`
	DueComplete           bool                       `json:"dueComplete"`
	LastUpdatedByAI       bool                       `json:"lastUpdatedByAi"`
	Start                 *DateTime                  `json:"start"`
	ExternalSource        *string                    `json:"externalSource"`
	AttachmentsByType     CardBadgeAttachmentsByType `json:"attachmentsByType"`
	Location              bool                       `json:"location"`
	Votes                 int                        `json:"votes"`
	MaliciousAttachments  int                        `json:"maliciousAttachments"`
	ViewingMemberVoted    bool                       `json:"viewingMemberVoted"`
	Subscribed            bool                       `json:"subscribed"`
}

type CardBadgeAttachmentsByType struct {
	Trello struct {
		Board int `json:"board"`
		Card  int `json:"card"`
	} `json:"trello"`
}

type CardLabelResponse struct {
	ID             string `json:"id"`
	BoardID        string `json:"idBoard"`
	OrganizationID string `json:"idOrganization"`
	Name           string `json:"name"`
	NodeID         string `json:"nodeId"`
	Color          string `json:"color"`
	Uses           int    `json:"uses"`
}

type CardCheckItemStateResponse struct {
	ID    string `json:"idCheckItem"`
	State string `json:"state"`
}

type NewCardRequest struct {
	ListID string `json:"idList"`

	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"desc,omitempty"`
	Position    *string   `json:"pos,omitempty" validate:"newCardPosition"`
	Due         *DateTime `json:"due,omitempty"`
	Start       *DateTime `json:"start,omitempty"`
	DueComplete *bool     `json:"dueComplete,omitempty"`
	MemberIDs   []string  `json:"idMembers,omitempty" validate:"dive,trelloID"`
	LabelIDs    []string  `json:"idLabels,omitempty" validate:"dive,trelloID"`
	URLSource   *string   `json:"urlSource,omitempty" validate:"url"`
}
