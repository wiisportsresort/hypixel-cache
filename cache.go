package main

import (
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var playerCache *cache.Cache = cache.New(15*time.Minute, 30*time.Minute)
var uuidCache *cache.Cache = cache.New(60*time.Minute, 120*time.Minute)

var uuidRegexp *regexp.Regexp = regexp.MustCompile("^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$")
var usernameRegexp *regexp.Regexp = regexp.MustCompile("^[a-zA-Z0-9_]{3,16}$")

type CachedPlayer struct {
	Player    map[string]interface{}
	FetchedAt time.Time
}

func findCachedPlayer(c *gin.Context) {
	if c.Writer.Written() {
		c.Next()
		return
	}

	identifierType := c.Param("type")
	identifier := c.Param("identifier")

	var uuid string

	if identifierType == "uuid" {
		if !uuidRegexp.MatchString(identifier) {
			finish(c, 400, NewFailure("invalid uuid"))
			return
		}
		uuid = identifier
	} else if identifierType == "name" {
		if !usernameRegexp.MatchString(identifier) {
			finish(c, 400, NewFailure("invalid username"))
			return
		}

		if cachedUuid, ok := uuidCache.Get(identifier); ok {
			uuid = cachedUuid.(string)
		} else {
			c.Next()
			return
		}
	} else {
		finish(c, 400, NewFailure("invalid identifier type"))

		return
	}

	uuid = removeDashes(uuid)

	if cached, ok := playerCache.Get(uuid); ok {
		cachedPlayer := cached.(CachedPlayer)
		if cachedPlayer.Player != nil {
			finish(c, 200, NewSuccessPlayerFound(cachedPlayer, true))
		} else {
			finish(c, 200, NewSuccessNotFound(cachedPlayer.FetchedAt, false))
		}

	} else {
		c.Next()
	}
}
