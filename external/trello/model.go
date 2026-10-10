package trello

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Types

type StringInt int

func (i *StringInt) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to unmarshal to string")
	}
	d, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("unable to parse %s as int", s)
	}
	*i = StringInt(d)
	return nil
}

func (i StringInt) MarshalJSON() ([]byte, error) {
	s := strconv.Itoa(int(i))
	return json.Marshal(s)
}

type DateTime time.Time

func (d *DateTime) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to unmarshal date to string")
	}
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		return fmt.Errorf("failed to parse date %s: %w", s, err)
	}
	*d = DateTime(t)
	return nil
}

func (d *DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(*d)
	formatted := t.UTC().Format("2006-01-02T15:04:05.000Z")
	return json.Marshal(formatted)
}

func (d *DateTime) Value() time.Time {
	return time.Time(*d)
}

type KeywordString bool

func (k *KeywordString) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to unmarshal field to string")
	}
	switch s {
	case "enabled":
		*k = true
	case "disabled":
		*k = false
	default:
		return fmt.Errorf("unrecognised keyword %s", s)
	}
	return nil
}

func (k KeywordString) MarshalJSON() ([]byte, error) {
	switch k {
	case true:
		return json.Marshal("enabled")
	case false:
		return json.Marshal("disabled")
	default:
		return nil, fmt.Errorf("invalid value")
	}
}

type KeywordFloat float64

const (
	KeywordTop    KeywordFloat = 0
	KeywordBottom KeywordFloat = -1
)

func (k *KeywordFloat) UnmarshalJSON(b []byte) error {
	var s string
	var f float64

	if err := json.Unmarshal(b, &s); err != nil {
		if err := json.Unmarshal(b, &f); err != nil {
			return fmt.Errorf("unable to unmarshal to either string or int")
		}
		*k = KeywordFloat(f)
		return nil
	}

	switch s {
	case "top":
		*k = KeywordTop
	case "bottom":
		*k = KeywordBottom
	default:
		return fmt.Errorf("unknown string param: %s", s)
	}
	return nil
}

func (k KeywordFloat) MarshalJSON() ([]byte, error) {
	switch k {
	case KeywordTop:
		return json.Marshal("top")
	case KeywordBottom:
		return json.Marshal("bottom")
	default:
		f := float64(k)
		return json.Marshal(f)
	}
}

type CheckItemState bool

func (c *CheckItemState) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "incomplete":
		*c = false
	case "complete":
		*c = true
	default:
		return fmt.Errorf("unknown state: %s", s)
	}

	return nil
}

func (c CheckItemState) MarshalJSON() ([]byte, error) {
	if c {
		return json.Marshal("complete")
	}
	return json.Marshal("incomplete")
}

func (c CheckItemState) Value() bool {
	return bool(c)
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

func getLabelColors() []string {
	return []string{"green", "yellow", "orange", "red", "purple", "blue", "sky", "lime", "pink", "black", "green_dark", "yellow_dark", "orange_dark", "red_dark", "purple_dark", "blue_dark", "sky_dark", "lime_dark", "pink_dark", "black_dark", "green_light", "yellow_light", "orange_light", "red_light", "purple_light", "blue_light", "sky_light", "lime_light", "pink_light", "black_light"}
}

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
	Memberships        []BoardMemberResponse    `json:"memberships"`
	ShortLink          string                   `json:"shortLink"`
	Subscribed         bool                     `json:"subscribed"`
	PowerUps           []string                 `json:"powerUps"`
	LastActivity       *DateTime                `json:"dateLastActivity"`
	LastViewed         *DateTime                `json:"dateLastView"`
	IDTags             []string                 `json:"idTags"`
	DatePluginDisabled *DateTime                `json:"datePluginDisable"`
	CreationMethod     string                   `json:"creationMethod"`
	IXUpdate           StringInt                `json:"ixUpdate"`
	TemplateGallery    string                   `json:"templateGallery"`
	EnterpriseOwned    bool                     `json:"enterpriseOwned"`
}

type BoardPreferencesResponse struct {
	PermissionLevel        BoardPermissionLevel `json:"permissionLevel"`
	HideVotes              bool                 `json:"hideVotes"`
	Voting                 KeywordString        `json:"voting"`
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

type BoardMemberResponse struct {
	ID          string `json:"id"`
	MemberID    string `json:"idMember"`
	MemberType  string `json:"memberType"`
	Unconfirmed bool   `json:"unconfirmed"`
	Deactivated bool   `json:"deactivated"`
}

type BoardLabelResponse struct {
	ID      string `json:"id"`
	BoardID string `json:"idBoard"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Uses    int    `json:"uses"`
}

type NewBoardRequest struct {
	Name                 string  `json:"name" validate:"omitempty,min=1,max=16384"`
	DefaultLabels        *bool   `json:"defaultLabels,omitempty"`
	DefaultLists         *bool   `json:"defaultLists,omitempty"`
	Description          *string `json:"desc,omitempty" validate:"omitempty,min=0,max=16384"`
	OrganizationID       *string `json:"idOrganization,omitempty" validate:"omitempty,trelloID"`
	SourceBoardID        *string `json:"idBoardSource,omitempty" validate:"omitempty,trelloID"`
	KeepFromSource       *string `json:"keepFromSource,omitempty" validate:"omitempty,oneof=cards none"`
	PowerUps             *string `json:"powerUps,omitempty" validate:"omitempty,oneof=all calendar cardAging recap voting"`
	PrefsPermissionLevel *string `json:"prefs_permissionLevel,omitempty" validate:"omitempty,oneof=org private public"`
	PrefsVoting          *string `json:"prefs_voting,omitempty" validate:"omitempty,oneof=disabled members observers org public"`
	PrefsComments        *string `json:"prefs_comments,omitempty" validate:"omitempty,oneof=disabled members observers org public"`
	PrefsInvitations     *string `json:"prefs_invitations,omitempty" validate:"omitempty,oneof=members admins"`
	PrefsSelfJoin        *bool   `json:"prefs_selfJoin,omitempty"`
	PrefsCardCovers      *bool   `json:"prefs_cardCovers,omitempty"`
	PrefsBackground      *string `json:"prefs_background,omitempty" validate:"omitempty,oneof=blue orange green red purple pink lime sky grey"`
	PrefsCardAging       *string `json:"prefs_cardAging,omitempty" validate:"omitempty,oneof=pirate regular"`
}

type NewBoardLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color" validate:"labelColor"`
}

type UpdateBoardRequest struct {
	Name           *string `json:"name,omitempty" validate:"omitempty,min=1,max=16384"`
	Description    *string `json:"desc,omitempty" validate:"omitempty,max=16384"`
	Closed         *bool   `json:"closed,omitempty"`
	Subscribed     *string `json:"subscribed,omitempty" validate:"omitempty,trelloID"`
	OrganizationID *string `json:"idOrganization,omitempty"`
}

// Board Lists

type BoardListResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Closed     bool                `json:"closed"`
	Position   float64             `json:"pos"`
	SoftLimit  string              `json:"softLimit"`
	BoardID    string              `json:"idBoard"`
	Subscribed bool                `json:"subscribed"`
	Limits     BoardLimitsResponse `json:"limits"`
}

type NewBoardListRequest struct {
	Name     string        `json:"name"`
	Position *KeywordFloat `json:"pos,omitempty"`
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
	ListPosition           float64                      `json:"pos"`
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
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"desc,omitempty"`
	Position    *KeywordFloat `json:"pos,omitempty"`
	Due         *DateTime     `json:"due,omitempty"`
	Start       *DateTime     `json:"start,omitempty"`
	DueComplete *bool         `json:"dueComplete,omitempty"`
	MemberIDs   []string      `json:"idMembers,omitempty" validate:"omitempty,dive,trelloID"`
	LabelIDs    []string      `json:"idLabels,omitempty" validate:"omitempty,dive,trelloID"`
	URLSource   *string       `json:"urlSource,omitempty" validate:"omitempty,url"`
}

type UpdateCardRequest struct {
	CardID string `json:"id"`

	Name              *string       `json:"name,omitempty"`
	Description       *string       `json:"desc,omitempty"`
	Closed            *bool         `json:"closed,omitempty"`
	MemberIDs         []string      `json:"idMembers,omitempty" validate:"omitempty,dive,trelloID"`
	AttachmentCoverID *string       `json:"idAttachmentCover,omitempty" validate:"omitempty,trelloID"`
	ListID            *string       `json:"idList,omitempty" validate:"omitempty,trelloID"`
	LabelIDs          []string      `json:"idLabels,omitempty" validate:"omitempty,dive,trelloID"`
	BoardID           *string       `json:"idBoard,omitempty" validate:"omitempty,trelloID"`
	Position          *KeywordFloat `json:"pos,omitempty"`
	Due               *DateTime     `json:"due,omitempty"`
	DueComplete       *bool         `json:"dueComplete,omitempty"`
}

// Checklists

type ChecklistResponse struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	BoardID  string              `json:"idBoard"`
	CardID   string              `json:"idCard"`
	Position float64             `json:"pos"`
	Items    []CheckitemResponse `json:"checkItems"`
}

type CheckitemResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	NameData    CheckitemNameDataResponse `json:"nameData"`
	Position    float64                   `json:"pos"`
	State       CheckItemState            `json:"state"`
	Due         *DateTime                 `json:"due"`
	DueReminder *DateTime                 `json:"dueReminder"`
	MemberID    *string                   `json:"idMember"`
	ChecklistID string                    `json:"idChecklist"`
}

type CheckitemNameDataResponse struct {
	Emoji any `json:"emoji"`
}

type NewChecklistRequest struct {
	CardID            *string       `json:"idCard,omitempty" validate:"omitempty,trelloID"`
	Name              *string       `json:"name,omitempty"`
	Position          *KeywordFloat `json:"pos,omitempty"`
	SourceChecklistID *string       `json:"idChecklistSource,omitempty" validate:"omitempty,trelloID"`
}

type NewCheckitemRequest struct {
	Name        string        `json:"name"`
	Position    *KeywordFloat `json:"pos,omitempty"`
	Checked     *bool         `json:"checked,omitempty"`
	Due         *DateTime     `json:"due,omitempty"`
	DueReminder *DateTime     `json:"dueReminder,omitempty"`
	MemberID    *string       `json:"idMember,omitempty" validate:"omitempty,trelloID"`
}

type UpdateCheckitemRequest struct {
	Name        *string         `json:"name,omitempty"`
	State       *CheckItemState `json:"state,omitempty"`
	ChecklistID *string         `json:"idChecklist,omitempty" validate:"omitempty,trelloID"`
	Position    *KeywordFloat   `json:"pos,omitempty"`
	Due         *DateTime       `json:"due,omitempty"`
	DueReminder *DateTime       `json:"dueReminder,omitempty"`
	MemberID    *string         `json:"idMember,omitempty" validate:"omitempty,trelloID"`
}
