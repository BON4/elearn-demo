package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type CourseServiceClient struct {
	BaseURL string
	Client  *http.Client
}

type CourseDTO struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (c *CourseDTO) toDomain() *domain.CourseRM {
	return &domain.CourseRM{
		ID:     c.ID,
		Status: c.Status,
	}
}

func NewCourseServiceClient(baseURL string) *CourseServiceClient {
	return &CourseServiceClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *CourseServiceClient) GetCourse(ctx context.Context, courseID uuid.UUID) (*domain.CourseRM, error) {
	path, err := url.JoinPath(c.BaseURL, "courses", courseID.String())
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(err, fmt.Sprintf("failed to call: %s", path))
		return nil, fmt.Errorf("course with id: %s not found, status: %d", courseID, resp.StatusCode)
	}

	var dto CourseDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	return dto.toDomain(), nil
}
