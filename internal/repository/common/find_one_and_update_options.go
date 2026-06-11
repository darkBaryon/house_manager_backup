package common

import "go.mongodb.org/mongo-driver/v2/mongo/options"

// FindOneAndUpdateOption configures typed FindOneAndUpdate operations.
type FindOneAndUpdateOption func(*findOneAndUpdateOptions)

type findOneAndUpdateOptions struct {
	returnAfter bool
}

// ReturnAfter makes FindOneAndUpdateBy return the updated document.
func ReturnAfter() FindOneAndUpdateOption {
	return func(opts *findOneAndUpdateOptions) {
		opts.returnAfter = true
	}
}

func buildFindOneAndUpdateOptions(opts ...FindOneAndUpdateOption) []options.Lister[options.FindOneAndUpdateOptions] {
	if len(opts) == 0 {
		return nil
	}

	typed := findOneAndUpdateOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&typed)
		}
	}

	builder := options.FindOneAndUpdate()
	if typed.returnAfter {
		builder.SetReturnDocument(options.After)
	}
	return []options.Lister[options.FindOneAndUpdateOptions]{builder}
}
