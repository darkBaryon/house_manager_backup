package publishaccess

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"
	"house-manager/pkg/session"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service owns the publish-side relation between a listing and the staff/user
// principal that may maintain it.
type Service struct {
	listingRepo hpdListingRepository
	entrustRepo hpdEntrustRelationRepository
}

func NewService(
	hpdListingRepo *repohpd.ListingRepository,
	hpdEntrustRelationRepo *repohpd.EntrustRelationRepository,
) *Service {
	return &Service{
		listingRepo: hpdListingRepo,
		entrustRepo: hpdEntrustRelationRepo,
	}
}

func (s *Service) FindListingBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error) {
	if s == nil || s.listingRepo == nil {
		return nil, databasef("find hpd listing by source: listing repo is nil")
	}
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, invalidParamf("source_type and source_id are required")
	}
	listing, err := s.listingRepo.FindBySource(ctx, sourceType, sourceID)
	if err != nil {
		return nil, databasef("find hpd listing by source: %w", err)
	}
	return listing, nil
}

func (s *Service) UpsertEntrustForPrincipal(ctx context.Context, listingID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdEntrustRelation, error) {
	if s == nil || s.entrustRepo == nil {
		return nil, databasef("upsert hpd entrust relation: entrust relation repo is nil")
	}
	relation, err := entrustRelationForPrincipal(listingID, principal)
	if err != nil {
		return nil, err
	}
	result, err := s.entrustRepo.UpsertActiveByListingID(ctx, relation)
	if err != nil {
		return nil, databasef("upsert hpd entrust relation: %w", err)
	}
	return result, nil
}

func (s *Service) ListAccessibleListings(ctx context.Context, principal session.Principal) ([]hpdmodel.HpdListing, error) {
	if s == nil || s.listingRepo == nil || s.entrustRepo == nil {
		return nil, databasef("list accessible hpd listings: repo is nil")
	}
	listingIDs, err := s.listingIDsForPrincipal(ctx, principal)
	if err != nil {
		return nil, err
	}
	if len(listingIDs) == 0 {
		return []hpdmodel.HpdListing{}, nil
	}
	listings, err := s.listingRepo.ListByIDs(ctx, listingIDs)
	if err != nil {
		return nil, databasef("list accessible hpd listings: %w", err)
	}
	return listings, nil
}

func (s *Service) CanAccessSourceForPrincipal(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID, principal session.Principal) (bool, error) {
	listing, err := s.FindListingBySource(ctx, sourceType, sourceID)
	if err != nil {
		return false, err
	}
	if listing == nil {
		return false, nil
	}
	return s.CanAccessListingForPrincipal(ctx, listing.ID, principal)
}

func (s *Service) CanAccessListingForPrincipal(ctx context.Context, listingID bson.ObjectID, principal session.Principal) (bool, error) {
	if s == nil || s.entrustRepo == nil {
		return false, databasef("can access hpd listing: entrust relation repo is nil")
	}
	staffID, ownerPhone, err := accessKeysForPrincipal(principal)
	if err != nil {
		return false, err
	}
	allowed, err := s.entrustRepo.CanAccessListing(ctx, listingID, staffID, ownerPhone)
	if err != nil {
		return false, databasef("can access hpd listing: %w", err)
	}
	return allowed, nil
}

func (s *Service) listingIDsForPrincipal(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error) {
	staffID, ownerPhone, err := accessKeysForPrincipal(principal)
	if err != nil {
		return nil, err
	}
	switch {
	case !staffID.IsZero():
		ids, err := s.entrustRepo.ListActiveListingIDsByStaff(ctx, staffID)
		if err != nil {
			return nil, databasef("list active hpd listing IDs by staff: %w", err)
		}
		return ids, nil
	case ownerPhone != "":
		ids, err := s.entrustRepo.ListActiveListingIDsByOwnerPhone(ctx, ownerPhone)
		if err != nil {
			return nil, databasef("list active hpd listing IDs by owner phone: %w", err)
		}
		return ids, nil
	default:
		return nil, invalidParamf("principal has no access key")
	}
}

func entrustRelationForPrincipal(listingID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdEntrustRelation, error) {
	if listingID.IsZero() {
		return nil, invalidParamf("listing_id is required")
	}
	staffID, ownerPhone, err := accessKeysForPrincipal(principal)
	if err != nil {
		return nil, err
	}
	relation := &hpdmodel.HpdEntrustRelation{
		ListingID:      listingID,
		RelationStatus: hpdmodel.HpdRelationStatusActive,
	}
	switch principal.PrincipalType {
	case session.PrincipalTypeStaff:
		relation.MaintainerStaffID = staffID
		relation.ServiceStaffID = staffID
	case session.PrincipalTypeUser:
		relation.OwnerPhone = ownerPhone
	default:
		return nil, invalidParamf("principal_type is invalid")
	}
	return relation, nil
}

func accessKeysForPrincipal(principal session.Principal) (bson.ObjectID, string, error) {
	if err := principal.Validate(); err != nil {
		return bson.NilObjectID, "", invalidParamf("principal is invalid: %w", err)
	}
	if principal.Terminal != session.TerminalPublish {
		return bson.NilObjectID, "", invalidParamf("principal terminal must be publish")
	}
	switch principal.PrincipalType {
	case session.PrincipalTypeStaff:
		staffID, err := bson.ObjectIDFromHex(principal.PrincipalID)
		if err != nil {
			return bson.NilObjectID, "", invalidParamf("staff principal_id is invalid: %w", err)
		}
		return staffID, "", nil
	case session.PrincipalTypeUser:
		ownerPhone := strings.TrimSpace(principal.Phone)
		if ownerPhone == "" {
			return bson.NilObjectID, "", invalidParamf("user principal phone is required")
		}
		return bson.NilObjectID, ownerPhone, nil
	default:
		return bson.NilObjectID, "", invalidParamf("principal_type is invalid")
	}
}
