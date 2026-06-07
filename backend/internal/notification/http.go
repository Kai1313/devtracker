package notification

import "context"

func CreateNotification(ctx context.Context, service *Service, input CreateInput) error {
	if service == nil {
		return nil
	}

	return service.Create(ctx, input)
}
