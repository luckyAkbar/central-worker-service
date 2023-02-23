package model

import "context"

type SubscriptionType string

var (
	SubscriptionTypeMeme SubscriptionType = "meme"
)

type SubscriptionChannel string

var (
	SubscriptionChannelTelegram SubscriptionChannel = "telegram"
)

type Subscription struct {
	ID              string              `json:"id"`
	Type            SubscriptionType    `json:"type"`
	Channel         SubscriptionChannel `json:"channel"`
	UserReferenceID string              `json:"user_reference_id"`
}

type SubscriptionUsecase interface {
	Create(ctx context.Context, subscription *Subscription) UsecaseError
	FindSubscriptions(ctx context.Context, limit, offset int) ([]Subscription, UsecaseError)
	FindSubscription(ctx context.Context, subType SubscriptionType, channel SubscriptionChannel, userID string) (*Subscription, UsecaseError)
}

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *Subscription) error
	FindSubscriptions(ctx context.Context, limit, offset int) ([]Subscription, error)
	FindSubscription(ctx context.Context, subType SubscriptionType, channel SubscriptionChannel, userID string) (*Subscription, error)
}
