package postgres

import (
	"context"
	"fmt"

	"github.com/global-news/country-service/internal/domain"
	types "github.com/global-news/shared-types"
	"github.com/jmoiron/sqlx"
)

type countryRepository struct {
	db *sqlx.DB
}

func NewCountryRepository(db *sqlx.DB) domain.CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) Fetch(ctx context.Context, filter domain.CountryFilter) ([]types.Country, error) {
	var countries []types.Country
	query := `SELECT * FROM countries WHERE 1=1`
	args := []interface{}{}
	argId := 1

	if filter.Region != "" {
		query += fmt.Sprintf(" AND region = $%d", argId)
		args = append(args, filter.Region)
		argId++
	}

	if filter.Query != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argId)
		args = append(args, "%"+filter.Query+"%")
		argId++
	}

	if filter.SortBy == "risk_desc" {
		query += " ORDER BY risk_score DESC"
	} else if filter.SortBy == "risk_asc" {
		query += " ORDER BY risk_score ASC"
	} else if filter.SortBy == "sentiment_desc" {
		query += " ORDER BY sentiment DESC"
	} else if filter.SortBy == "sentiment_asc" {
		query += " ORDER BY sentiment ASC"
	} else {
		query += " ORDER BY name ASC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, filter.Limit, filter.Offset)

	err := r.db.SelectContext(ctx, &countries, query, args...)
	return countries, err
}

func (r *countryRepository) GetByCode(ctx context.Context, code string) (*types.Country, error) {
	var country types.Country
	query := `SELECT * FROM countries WHERE code = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &country, query, code)
	if err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) GetRankings(ctx context.Context, criteria string, limit int) ([]types.Country, error) {
	var countries []types.Country
	query := `SELECT * FROM countries ORDER BY `

	if criteria == "risk" {
		query += "risk_score DESC"
	} else if criteria == "sentiment" {
		query += "sentiment ASC"
	} else {
		query += "risk_score DESC"
	}

	query += ` LIMIT $1`
	err := r.db.SelectContext(ctx, &countries, query, limit)
	return countries, err
}
