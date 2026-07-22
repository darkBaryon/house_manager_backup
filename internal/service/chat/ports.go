package chat

import "context"

type AIResponder interface {
	Respond(ctx context.Context, input AIRespondInput) (AIRespondOutput, error)
}
