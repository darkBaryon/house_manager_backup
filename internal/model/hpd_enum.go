package model

type HpdSourceType string

const (
	HpdSourceTypeCentralizedRoom   HpdSourceType = "centralized_room"
	HpdSourceTypeDecentralizedRoom HpdSourceType = "decentralized_room"
)

type HpdAssetMode string

const (
	HpdAssetModeCentralized   HpdAssetMode = "centralized"
	HpdAssetModeDecentralized HpdAssetMode = "decentralized"
)

type HpdListingStatus int

const (
	HpdListingStatusUnspecified HpdListingStatus = 0
	HpdListingStatusDraft       HpdListingStatus = 1
	HpdListingStatusPending     HpdListingStatus = 2
	HpdListingStatusPublished   HpdListingStatus = 3
	HpdListingStatusOffline     HpdListingStatus = 4
	HpdListingStatusRented      HpdListingStatus = 5
	HpdListingStatusDeleted     HpdListingStatus = -1
)

type HpdOnlineStatus int

const (
	HpdOnlineStatusNo  HpdOnlineStatus = 0
	HpdOnlineStatusYes HpdOnlineStatus = 1
)

var validHpdSourceTypes = map[HpdSourceType]struct{}{
	HpdSourceTypeCentralizedRoom:   {},
	HpdSourceTypeDecentralizedRoom: {},
}

var validHpdAssetModes = map[HpdAssetMode]struct{}{
	HpdAssetModeCentralized:   {},
	HpdAssetModeDecentralized: {},
}

var validHpdListingStatuses = map[HpdListingStatus]struct{}{
	HpdListingStatusUnspecified: {},
	HpdListingStatusDraft:       {},
	HpdListingStatusPending:     {},
	HpdListingStatusPublished:   {},
	HpdListingStatusOffline:     {},
	HpdListingStatusRented:      {},
	HpdListingStatusDeleted:     {},
}

var validHpdOnlineStatuses = map[HpdOnlineStatus]struct{}{
	HpdOnlineStatusNo:  {},
	HpdOnlineStatusYes: {},
}

func (v HpdSourceType) Valid() bool {
	_, ok := validHpdSourceTypes[v]
	return ok
}

func (v HpdAssetMode) Valid() bool {
	_, ok := validHpdAssetModes[v]
	return ok
}

func (v HpdListingStatus) Valid() bool {
	_, ok := validHpdListingStatuses[v]
	return ok
}

func (v HpdOnlineStatus) Valid() bool {
	_, ok := validHpdOnlineStatuses[v]
	return ok
}

func HpdAssetModeForSourceType(sourceType HpdSourceType) (HpdAssetMode, bool) {
	switch sourceType {
	case HpdSourceTypeCentralizedRoom:
		return HpdAssetModeCentralized, true
	case HpdSourceTypeDecentralizedRoom:
		return HpdAssetModeDecentralized, true
	default:
		return "", false
	}
}

func HpdSourceAssetModeMatch(sourceType HpdSourceType, assetMode HpdAssetMode) bool {
	expected, ok := HpdAssetModeForSourceType(sourceType)
	return ok && expected == assetMode
}

func HpdMiniappOnlineStatus(listingStatus HpdListingStatus, roomStatus RoomStatus) HpdOnlineStatus {
	if listingStatus == HpdListingStatusPublished && roomStatus == RoomStatusAvailable {
		return HpdOnlineStatusYes
	}
	return HpdOnlineStatusNo
}
