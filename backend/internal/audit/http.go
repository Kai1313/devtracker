package audit

import "github.com/gofiber/fiber/v2"

func RecordHTTPRequest(c *fiber.Ctx, service *Service, input RecordInput) error {
	if service == nil {
		return nil
	}

	input.IPAddress = c.IP()
	input.UserAgent = c.Get(fiber.HeaderUserAgent)
	return service.Record(c.UserContext(), input)
}
