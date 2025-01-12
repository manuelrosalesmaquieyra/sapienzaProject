package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.PUT("/users/{username}", rt.setMyUserName)
	rt.router.POST("/users/{username}/photo", rt.setMyPhoto)
	rt.router.GET("/users/{username}/conversations", rt.getUserConversations)

	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	// Rutas de conversaciones
	rt.router.GET("/conversations/:conversationId/messages", rt.getConversationMessages)
	rt.router.POST("/conversations/:conversationId/messages", rt.sendMessage)

	return rt.router
}
