package instance

import "context"

type IInstanceRepository interface {
	Exists(ctx context.Context) (bool, error)
	CreateIfAbsent(ctx context.Context, data *Instance) error
	List(ctx context.Context) ([]*Instance, error)
	Update(ctx context.Context, data *Instance) error
}
