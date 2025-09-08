package examples

import "time"

// AnalyticsStats represents analytics data
type AnalyticsStats struct {
	PageViews      int64
	UniqueVisitors int64
	AvgSessionTime string
	TopPages       []PageStats
	LastUpdated    time.Time
}

// PageStats represents statistics for a single page
type PageStats struct {
	Path  string
	Views int64
}

// BlogPost represents a blog post
type BlogPost struct {
	ID          int
	Title       string
	Slug        string
	Content     string
	Excerpt     string
	Author      string
	PublishedAt time.Time
	UpdatedAt   time.Time
	Tags        []string
	Category    string
}