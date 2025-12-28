package graphql

import (
	"database/sql"
	"github.com/apimgr/weather/src/server/handler"
	"github.com/apimgr/weather/src/server/service"
)

// This file will NOT be regenerated automatically by gqlgen.
// Resolver is the root GraphQL resolver.
// It holds references to all services and handlers needed for resolving queries and mutations.
type Resolver struct {
	// Database connections
	ServerDB *sql.DB
	UsersDB  *sql.DB

	// Services
	WeatherService *service.WeatherService

	// Handlers (we'll use their logic in resolvers)
	APIHandler          *handlers.APIHandler
	AuthHandler         *handlers.AuthHandler
	LocationHandler     *handlers.LocationHandler
	NotificationHandler *handlers.NotificationHandler
	AdminHandler        *handlers.AdminHandler
	SettingsHandler     *handlers.AdminSettingsHandler
	SchedulerHandler    *handlers.SchedulerHandler
	ChannelHandler      *handlers.ChannelHandler
	EarthquakeHandler   *handlers.EarthquakeHandler
	HurricaneHandler    *handlers.HurricaneHandler
	SevereWeatherHandler *handlers.SevereWeatherHandler
	MoonHandler         *handlers.MoonHandler
}

// NewResolver creates a new root resolver with all dependencies
func NewResolver(
	serverDB, usersDB *sql.DB,
	weatherService *service.WeatherService,
	apiHandler *handlers.APIHandler,
	authHandler *handlers.AuthHandler,
	locationHandler *handlers.LocationHandler,
	notificationHandler *handlers.NotificationHandler,
	adminHandler *handlers.AdminHandler,
	settingsHandler *handlers.AdminSettingsHandler,
	schedulerHandler *handlers.SchedulerHandler,
	channelHandler *handlers.ChannelHandler,
	earthquakeHandler *handlers.EarthquakeHandler,
	hurricaneHandler *handlers.HurricaneHandler,
	severeWeatherHandler *handlers.SevereWeatherHandler,
	moonHandler *handlers.MoonHandler,
) *Resolver {
	return &Resolver{
		ServerDB:             serverDB,
		UsersDB:              usersDB,
		WeatherService:       weatherService,
		APIHandler:           apiHandler,
		AuthHandler:          authHandler,
		LocationHandler:      locationHandler,
		NotificationHandler:  notificationHandler,
		AdminHandler:         adminHandler,
		SettingsHandler:      settingsHandler,
		SchedulerHandler:     schedulerHandler,
		ChannelHandler:       channelHandler,
		EarthquakeHandler:    earthquakeHandler,
		HurricaneHandler:     hurricaneHandler,
		SevereWeatherHandler: severeWeatherHandler,
		MoonHandler:          moonHandler,
	}
}
