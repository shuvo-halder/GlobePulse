package types

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Role         Role      `json:"role" db:"role"`
	IsVerified   bool      `json:"is_verified" db:"is_verified"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Country struct {
	CountryCode      string    `json:"country_code" db:"country_code"`
	Name             string    `json:"name" db:"name"`
	Region           string    `json:"region" db:"region"`
	Population       int64     `json:"population" db:"population"`
	RiskScore        float64   `json:"risk_score" db:"risk_score"`
	Sentiment        float64   `json:"sentiment" db:"sentiment"`
	TrendingTopics   string    `json:"trending_topics" db:"trending_topics"`       // JSON representation
	RelatedCountries string    `json:"related_countries" db:"related_countries"` // JSON representation
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type NewsArticle struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Summary     string    `json:"summary" db:"summary"`
	Content     string    `json:"content" db:"content"`
	URL         string    `json:"url" db:"url"`
	SourceID    string    `json:"source_id" db:"source_id"`
	AuthorID    string    `json:"author_id" db:"author_id"`
	Language    string    `json:"language" db:"language"`
	CountryCode string    `json:"country_code" db:"country_code"`
	Sentiment   float64   `json:"sentiment" db:"sentiment"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type EntitySchema struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
}

type TopicSchema struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type AIAnalysisResult struct {
	NewsID          string         `json:"news_id"`
	Summary         string         `json:"summary"`
	SentimentScore  float64        `json:"sentiment_score"`
	SentimentLabel  string         `json:"sentiment_label"`
	Entities        []EntitySchema `json:"entities"`
	Topics          []TopicSchema  `json:"topics"`
	Countries       []string       `json:"countries"`
	EventType       string         `json:"event_type"`
	ImportanceScore float64        `json:"importance_score"`
	AIInsights      string         `json:"ai_insights"`
	ProcessedAt     time.Time      `json:"processed_at"`
}

type EventType string

const (
	EventNewsPublished EventType = "news_published"
	EventUserActivity  EventType = "user_activity"
)

type AnalyticsEvent struct {
	ID          string                 `json:"id" db:"id"`
	EventType   EventType              `json:"event_type" db:"event_type"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	CountryCode string                 `json:"country_code" db:"country_code"`
	Sentiment   float64                `json:"sentiment" db:"sentiment"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
}

type CountryMetrics struct {
	CountryCode   string    `json:"country_code" db:"country_code"`
	Date          time.Time `json:"date" db:"date"`
	TotalEvents   int64     `json:"total_events" db:"total_events"`
	AvgSentiment  float64   `json:"avg_sentiment" db:"avg_sentiment"`
	TrendingScore float64   `json:"trending_score" db:"trending_score"`
}

type HeatmapData struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Intensity float64 `json:"intensity"`
}
