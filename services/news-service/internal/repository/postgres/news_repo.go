package postgres

import (
	"context"
	"fmt"

	"github.com/global-news/news-service/internal/domain"
	types "github.com/global-news/shared-types"
	"github.com/jmoiron/sqlx"
)

type newsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) domain.NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Fetch(ctx context.Context, filter domain.NewsFilter) ([]types.NewsArticle, error) {
	var articles []types.NewsArticle

	query := `SELECT id, title, summary, content, url, source_id, country_code, sentiment, published_at, created_at, updated_at 
	          FROM news_articles WHERE 1=1`
	args := []interface{}{}
	argId := 1

	if filter.CountryCode != "" {
		query += fmt.Sprintf(" AND country_code = $%d", argId)
		args = append(args, filter.CountryCode)
		argId++
	}

	if filter.Topic != "" {
		// Assuming topics is a related table or JSON column, but keeping it simple for demonstration
		query += fmt.Sprintf(" AND id IN (SELECT article_id FROM article_topics WHERE topic = $%d)", argId)
		args = append(args, filter.Topic)
		argId++
	}

	if filter.Query != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR summary ILIKE $%d)", argId, argId)
		searchQuery := "%" + filter.Query + "%"
		args = append(args, searchQuery)
		argId++
	}

	query += fmt.Sprintf(" ORDER BY published_at DESC LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, filter.Limit, filter.Offset)

	err := r.db.SelectContext(ctx, &articles, query, args...)
	return articles, err
}

func (r *newsRepository) GetByID(ctx context.Context, id string) (*types.NewsArticle, error) {
	var article types.NewsArticle
	query := `SELECT * FROM news_articles WHERE id = $1`
	err := r.db.GetContext(ctx, &article, query, id)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *newsRepository) Create(ctx context.Context, article *types.NewsArticle) error {
	query := `
		INSERT INTO news_articles (id, title, summary, content, url, source_id, country_code, sentiment, published_at) 
		VALUES (:id, :title, :summary, :content, :url, :source_id, :country_code, :sentiment, :published_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, article)
	return err
}
