package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// EducatorProfileDAO represents educator profile data from database
type EducatorProfileDAO struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Bio                sql.NullString
	Tagline            sql.NullString
	ExpertiseAreas     pq.StringArray
	Credentials        sql.NullString // JSON array stored as string
	SocialLinks        []byte         // JSONB stored as bytes
	ProfileBannerURL   sql.NullString
	IsFeatured         bool
	IsVerified         bool
	TotalSubscribers   int
	TotalQuizzes       int
	AverageRating      float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// EducatorProfileAPI represents educator profile data for API responses
type EducatorProfileAPI struct {
	ID                 uuid.UUID         `json:"id"`
	UserID             uuid.UUID         `json:"user_id"`
	Bio                string            `json:"bio,omitempty"`
	Tagline            string            `json:"tagline,omitempty"`
	ExpertiseAreas     []string          `json:"expertise_areas,omitempty"`
	Credentials        []string          `json:"credentials,omitempty"`
	SocialLinks        map[string]string `json:"social_links,omitempty"`
	ProfileBannerURL   string            `json:"profile_banner_url,omitempty"`
	IsFeatured         bool              `json:"is_featured"`
	IsVerified         bool              `json:"is_verified"`
	TotalSubscribers   int               `json:"total_subscribers"`
	TotalQuizzes       int               `json:"total_quizzes"`
	AverageRating      float64           `json:"average_rating"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

// ToAPI converts EducatorProfileDAO to EducatorProfileAPI
func (p *EducatorProfileDAO) ToAPI() EducatorProfileAPI {
	api := EducatorProfileAPI{
		ID:               p.ID,
		UserID:           p.UserID,
		IsFeatured:       p.IsFeatured,
		IsVerified:       p.IsVerified,
		TotalSubscribers: p.TotalSubscribers,
		TotalQuizzes:     p.TotalQuizzes,
		AverageRating:    p.AverageRating,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
	}

	if p.Bio.Valid {
		api.Bio = p.Bio.String
	}

	if p.Tagline.Valid {
		api.Tagline = p.Tagline.String
	}

	if len(p.ExpertiseAreas) > 0 {
		api.ExpertiseAreas = p.ExpertiseAreas
	}

	if p.Credentials.Valid && p.Credentials.String != "" {
		var creds []string
		json.Unmarshal([]byte(p.Credentials.String), &creds)
		api.Credentials = creds
	}

	if len(p.SocialLinks) > 0 {
		var links map[string]string
		json.Unmarshal(p.SocialLinks, &links)
		api.SocialLinks = links
	}

	if p.ProfileBannerURL.Valid {
		api.ProfileBannerURL = p.ProfileBannerURL.String
	}

	return api
}

// CreateEducatorProfileRequest represents a request to create/update educator profile
type CreateEducatorProfileRequest struct {
	Bio              string            `json:"bio,omitempty"`
	Tagline          string            `json:"tagline,omitempty"`
	ExpertiseAreas   []string          `json:"expertise_areas,omitempty"`
	Credentials      []string          `json:"credentials,omitempty"`
	SocialLinks      map[string]string `json:"social_links,omitempty"`
	ProfileBannerURL string            `json:"profile_banner_url,omitempty"`
}

// UpdateEducatorProfileRequest represents a request to update educator profile
type UpdateEducatorProfileRequest struct {
	Bio              *string           `json:"bio,omitempty"`
	Tagline          *string           `json:"tagline,omitempty"`
	ExpertiseAreas   []string          `json:"expertise_areas,omitempty"`
	Credentials      []string          `json:"credentials,omitempty"`
	SocialLinks      map[string]string `json:"social_links,omitempty"`
	ProfileBannerURL *string           `json:"profile_banner_url,omitempty"`
}

// PromoteToEducatorRequest represents a request to promote to educator
type PromoteToEducatorRequest struct {
	Bio            string   `json:"bio,omitempty"`
	Tagline        string   `json:"tagline,omitempty"`
	ExpertiseAreas []string `json:"expertise_areas,omitempty"`
}

// EducatorListResponse represents a list of educators
type EducatorListResponse struct {
	Educators []EducatorWithUserInfo `json:"educators"`
	Total     int                    `json:"total"`
}

// EducatorWithUserInfo combines educator profile with user info
type EducatorWithUserInfo struct {
	UserID    uuid.UUID          `json:"user_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email,omitempty"`
	FullName  string             `json:"full_name,omitempty"`
	AvatarURL string             `json:"avatar_url,omitempty"`
	Profile   EducatorProfileAPI `json:"profile"`
}
