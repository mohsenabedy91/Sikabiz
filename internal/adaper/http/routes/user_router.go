package routes

import (
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/handler"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/middlewares"
)

func (r *Router) NewUserRouter(userHandler handler.UserHandler) *Router {
	v1 := r.Engine.Group(":language/v1", middlewares.LocaleMiddleware(r.trans))
	{
		user := v1.Group("users")
		{
			user.GET(":userID", userHandler.Get)
		}
	}

	return &Router{
		Engine: r.Engine,
		log:    r.log,
		conf:   r.conf,
		trans:  r.trans,
	}
}
