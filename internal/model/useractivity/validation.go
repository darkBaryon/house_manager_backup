package useractivity

import "fmt"

func (m *Favorite) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("favorite is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if m.ListingID.IsZero() {
		return fmt.Errorf("listingID is required")
	}
	return nil
}

func (m *History) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("history is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if m.ListingID.IsZero() {
		return fmt.Errorf("listingID is required")
	}
	if !ValidHistorySource(m.Source) {
		return fmt.Errorf("source is invalid")
	}
	if m.ViewedAt <= 0 {
		return fmt.Errorf("viewedAt is required")
	}
	return nil
}
