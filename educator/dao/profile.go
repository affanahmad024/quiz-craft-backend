package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/educator/model"
)

var db *sql.DB

// Init initializes the DAO with database connection
func Init(database *sql.DB) {
	db = database
}

// CreateProfile creates a new educator profile
func CreateProfile(ctx context.Context, userID uuid.UUID, req model.CreateEducatorProfileRequest) (*model.EducatorProfileDAO, error) {
	commons.Log(ctx, "[dao.CreateProfile [user_id=%s", userID)

	profile := &model.EducatorProfileDAO{
		ID:     uuid.New(),
		UserID: userID,
	}

	// Prepare optional fields
	var bio, tagline sql.NullString
	if req.Bio != "" {
		bio = sql.NullString{String: req.Bio, Valid: true}
	}
	if req.Tagline != "" {
		tagline = sql.NullString{String: req.Tagline, Valid: true}
	}

	expertiseAreas := pq.StringArray(req.ExpertiseAreas)

	var credentialsJSON sql.NullString
	if len(req.Credentials) > 0 {
		credsBytes, _ := json.Marshal(req.Credentials)
		credentialsJSON = sql.NullString{String: string(credsBytes), Valid: true}
	}

	socialLinksBytes, _ := json.Marshal(req.SocialLinks)
	if len(socialLinksBytes) == 0 || string(socialLinksBytes) == "null" {
		socialLinksBytes = []byte("{}")
	}

	var profileBannerURL sql.NullString
	if req.ProfileBannerURL != "" {
		profileBannerURL = sql.NullString{String: req.ProfileBannerURL, Valid: true}
	}

	qry := `INSERT INTO educator_profiles (
				id, user_id, bio, tagline, expertise_areas, credentials,
				social_links, profile_banner_url, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
			) RETURNING created_at, updated_at`

	err := db.QueryRowContext(ctx, qry,
		profile.ID,
		profile.UserID,
		bio,
		tagline,
		expertiseAreas,
		credentialsJSON,
		socialLinksBytes,
		profileBannerURL,
	).Scan(&profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateProfile::SQLErr [err=%v", err)
		return nil, err
	}

	// Populate the created profile fields
	profile.Bio = bio
	profile.Tagline = tagline
	profile.ExpertiseAreas = expertiseAreas
	profile.Credentials = credentialsJSON
	profile.SocialLinks = socialLinksBytes
	profile.ProfileBannerURL = profileBannerURL

	commons.Log(ctx, "[dao.CreateProfile::OK [profile_id=%s", profile.ID)
	return profile, nil
}

// FindByUserID finds educator profile by user ID
func FindByUserID(ctx context.Context, userID uuid.UUID) (*model.EducatorProfileDAO, error) {
	commons.Log(ctx, "[dao.FindByUserID [user_id=%s", userID)

	profile := &model.EducatorProfileDAO{}
	qry := `SELECT id, user_id, bio, tagline, expertise_areas, credentials,
			social_links, profile_banner_url, is_featured, is_verified,
			total_subscribers, total_quizzes, average_rating,
			created_at, updated_at
			FROM educator_profiles
			WHERE user_id = $1`

	err := db.QueryRowContext(ctx, qry, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Bio,
		&profile.Tagline,
		&profile.ExpertiseAreas,
		&profile.Credentials,
		&profile.SocialLinks,
		&profile.ProfileBannerURL,
		&profile.IsFeatured,
		&profile.IsVerified,
		&profile.TotalSubscribers,
		&profile.TotalQuizzes,
		&profile.AverageRating,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, commons.NotFound("educator-profile-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByUserID::SQLErr [user_id=%s [err=%v", userID, err)
		return nil, err
	}

	return profile, nil
}

// UpdateProfile updates educator profile fields
func UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateEducatorProfileRequest) error {
	commons.Log(ctx, "[dao.UpdateProfile [user_id=%s", userID)

	// Build dynamic update query
	query := `UPDATE educator_profiles SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.Bio != nil {
		query += `, bio = $` + strconv.Itoa(argIndex)
		args = append(args, *req.Bio)
		argIndex++
	}

	if req.Tagline != nil {
		query += `, tagline = $` + strconv.Itoa(argIndex)
		args = append(args, *req.Tagline)
		argIndex++
	}

	if req.ExpertiseAreas != nil {
		query += `, expertise_areas = $` + strconv.Itoa(argIndex)
		args = append(args, pq.StringArray(req.ExpertiseAreas))
		argIndex++
	}

	if req.Credentials != nil {
		credsBytes, _ := json.Marshal(req.Credentials)
		query += `, credentials = $` + strconv.Itoa(argIndex)
		args = append(args, string(credsBytes))
		argIndex++
	}

	if req.SocialLinks != nil {
		linksBytes, _ := json.Marshal(req.SocialLinks)
		query += `, social_links = $` + strconv.Itoa(argIndex)
		args = append(args, linksBytes)
		argIndex++
	}

	if req.ProfileBannerURL != nil {
		query += `, profile_banner_url = $` + strconv.Itoa(argIndex)
		args = append(args, *req.ProfileBannerURL)
		argIndex++
	}

	query += ` WHERE user_id = $` + strconv.Itoa(argIndex)
	args = append(args, userID)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateProfile::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("educator-profile-not-found")
	}

	commons.Log(ctx, "[dao.UpdateProfile::OK [user_id=%s", userID)
	return nil
}

// ListEducators lists all educators with pagination
func ListEducators(ctx context.Context, limit, offset int) ([]model.EducatorProfileDAO, int, error) {
	commons.Log(ctx, "[dao.ListEducators [limit=%d [offset=%d", limit, offset)

	// Get total count
	var total int
	qry := `SELECT COUNT(*) FROM educator_profiles`
	err := db.QueryRowContext(ctx, qry).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListEducators::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Get educators
	profiles := []model.EducatorProfileDAO{}
	qry = `SELECT id, user_id, bio, tagline, expertise_areas, credentials,
			social_links, profile_banner_url, is_featured, is_verified,
			total_subscribers, total_quizzes, average_rating,
			created_at, updated_at
			FROM educator_profiles
			ORDER BY total_subscribers DESC, created_at DESC
			LIMIT $1 OFFSET $2`

	rows, err := db.QueryContext(ctx, qry, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.ListEducators::SQLErr [err=%v", err)
		return profiles, total, err
	}
	defer rows.Close()

	for rows.Next() {
		profile := model.EducatorProfileDAO{}
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Bio,
			&profile.Tagline,
			&profile.ExpertiseAreas,
			&profile.Credentials,
			&profile.SocialLinks,
			&profile.ProfileBannerURL,
			&profile.IsFeatured,
			&profile.IsVerified,
			&profile.TotalSubscribers,
			&profile.TotalQuizzes,
			&profile.AverageRating,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListEducators::ScanErr [err=%v", err)
			continue
		}
		profiles = append(profiles, profile)
	}

	commons.Log(ctx, "[dao.ListEducators::OK [count=%d [total=%d", len(profiles), total)
	return profiles, total, nil
}

// SearchEducators searches educators by name or expertise
func SearchEducators(ctx context.Context, query string, limit, offset int) ([]model.EducatorProfileDAO, int, error) {
	commons.Log(ctx, "[dao.SearchEducators [query=%s", query)

	searchPattern := "%" + query + "%"

	// Get total count
	var total int
	qry := `SELECT COUNT(*)
			FROM educator_profiles ep
			JOIN users u ON ep.user_id = u.id
			WHERE u.username ILIKE $1
			   OR u.full_name ILIKE $1
			   OR ep.tagline ILIKE $1
			   OR $2 = ANY(ep.expertise_areas)`

	err := db.QueryRowContext(ctx, qry, searchPattern, query).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.SearchEducators::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Get educators
	profiles := []model.EducatorProfileDAO{}
	qry = `SELECT ep.id, ep.user_id, ep.bio, ep.tagline, ep.expertise_areas, ep.credentials,
			ep.social_links, ep.profile_banner_url, ep.is_featured, ep.is_verified,
			ep.total_subscribers, ep.total_quizzes, ep.average_rating,
			ep.created_at, ep.updated_at
			FROM educator_profiles ep
			JOIN users u ON ep.user_id = u.id
			WHERE u.username ILIKE $1
			   OR u.full_name ILIKE $1
			   OR ep.tagline ILIKE $1
			   OR $2 = ANY(ep.expertise_areas)
			ORDER BY ep.total_subscribers DESC, ep.created_at DESC
			LIMIT $3 OFFSET $4`

	rows, err := db.QueryContext(ctx, qry, searchPattern, query, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.SearchEducators::SQLErr [err=%v", err)
		return profiles, total, err
	}
	defer rows.Close()

	for rows.Next() {
		profile := model.EducatorProfileDAO{}
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Bio,
			&profile.Tagline,
			&profile.ExpertiseAreas,
			&profile.Credentials,
			&profile.SocialLinks,
			&profile.ProfileBannerURL,
			&profile.IsFeatured,
			&profile.IsVerified,
			&profile.TotalSubscribers,
			&profile.TotalQuizzes,
			&profile.AverageRating,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.SearchEducators::ScanErr [err=%v", err)
			continue
		}
		profiles = append(profiles, profile)
	}

	commons.Log(ctx, "[dao.SearchEducators::OK [count=%d [total=%d", len(profiles), total)
	return profiles, total, nil
}

// GetFeaturedEducators gets featured educators
func GetFeaturedEducators(ctx context.Context, limit int) ([]model.EducatorProfileDAO, error) {
	commons.Log(ctx, "[dao.GetFeaturedEducators [limit=%d", limit)

	profiles := []model.EducatorProfileDAO{}
	qry := `SELECT id, user_id, bio, tagline, expertise_areas, credentials,
			social_links, profile_banner_url, is_featured, is_verified,
			total_subscribers, total_quizzes, average_rating,
			created_at, updated_at
			FROM educator_profiles
			WHERE is_featured = true
			ORDER BY total_subscribers DESC
			LIMIT $1`

	rows, err := db.QueryContext(ctx, qry, limit)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFeaturedEducators::SQLErr [err=%v", err)
		return profiles, err
	}
	defer rows.Close()

	for rows.Next() {
		profile := model.EducatorProfileDAO{}
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Bio,
			&profile.Tagline,
			&profile.ExpertiseAreas,
			&profile.Credentials,
			&profile.SocialLinks,
			&profile.ProfileBannerURL,
			&profile.IsFeatured,
			&profile.IsVerified,
			&profile.TotalSubscribers,
			&profile.TotalQuizzes,
			&profile.AverageRating,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetFeaturedEducators::ScanErr [err=%v", err)
			continue
		}
		profiles = append(profiles, profile)
	}

	commons.Log(ctx, "[dao.GetFeaturedEducators::OK [count=%d", len(profiles))
	return profiles, nil
}

// DeleteProfile deletes educator profile
func DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.DeleteProfile [user_id=%s", userID)

	qry := `DELETE FROM educator_profiles WHERE user_id = $1`

	result, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteProfile::SQLErr [user_id=%s [err=%v", userID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("educator-profile-not-found")
	}

	commons.Log(ctx, "[dao.DeleteProfile::OK [user_id=%s", userID)
	return nil
}
