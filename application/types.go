package application

import "time"

type MigrateConfigModel struct {
	Key            string         `json:"key"`
	Value          string         `json:"value"`
	Type           ConfigType     `json:"type"`
	Description    string         `json:"description"`
	AdminOnly      bool           `json:"admin_only"`
	Category       ConfigCategory `json:"category"`
	ReleaseVersion string         `json:"release_version"`
}
type ConfigType string

const (
	ConfigTypeString = ConfigType("string")
	ConfigTypeNumber = ConfigType("number")
	ConfigTypeObject = ConfigType("object")
)

type ConfigCategory string

const (
	ConfigCategoryApplications = ConfigCategory("applications")
	ConfigCategoryTokens       = ConfigCategory("tokens")
	ConfigCategoryContent      = ConfigCategory("content")
	ConfigCategoryAd           = ConfigCategory("ad")
)

type ConfigModel struct {
	Key            string         `json:"key"`
	Value          string         `json:"value"`
	Type           ConfigType     `json:"type"`
	Description    string         `json:"description"`
	AdminOnly      bool           `json:"admin_only"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Category       ConfigCategory `json:"category"`
	ReleaseVersion string         `json:"release_version"`
}

type MigratorRequest struct {
	Configs map[string]MigrateConfigModel
}
