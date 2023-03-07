package model

import "context"

// SubscriptionType is a type that represent subscription type
type SubscriptionType string

// list subscription type
var (
	SubscriptionTypeMeme SubscriptionType = "meme"
)

// SubscriptionChannel is a type that represent subscription channel
type SubscriptionChannel string

// list subscription channel
var (
	SubscriptionChannelTelegram SubscriptionChannel = "telegram"
)

// Subscription is a model that represent the subscription
type Subscription struct {
	ID              string              `json:"id"`
	Type            SubscriptionType    `json:"type"`
	Channel         SubscriptionChannel `json:"channel"`
	UserReferenceID string              `json:"user_reference_id"`
}

// SubscriptionUsecase is a usecase that represent the subscription usecase
type SubscriptionUsecase interface {
	Create(ctx context.Context, subscription *Subscription) UsecaseError
	FindSubscriptions(ctx context.Context, limit, offset int) ([]Subscription, UsecaseError)
	FindSubscription(ctx context.Context, subType SubscriptionType, channel SubscriptionChannel, userID string) (*Subscription, UsecaseError)
	Delete(ctx context.Context, id string) UsecaseError
}

// SubscriptionRepository is a repository that represent the subscription repository
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *Subscription) error
	FindSubscriptions(ctx context.Context, limit, offset int) ([]Subscription, error)
	FindSubscription(ctx context.Context, subType SubscriptionType, channel SubscriptionChannel, userID string) (*Subscription, error)
	Delete(ctx context.Context, id string) error
}
