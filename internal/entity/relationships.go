package entity

type RelationshipType string

const (
	RelationshipMate    RelationshipType = "mate"
	RelationshipParent  RelationshipType = "parent"
	RelationshipChild   RelationshipType = "child"
	RelationshipSibling RelationshipType = "sibling"
	RelationshipFriend  RelationshipType = "friend"
	RelationshipRival   RelationshipType = "rival"
	RelationshipLove    RelationshipType = "love"
)

type EntityRelationship struct {
	OtherID   string           `json:"other_id"`
	Type      RelationshipType `json:"type"`
	SinceTick uint64           `json:"since_tick"`
}
