package internal

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"net/http"
	"project/generator"
)

type Service struct {
	db              *sqlx.DB
	postIDGenerator *generator.Generator
	userCache       *UserCache
}

func RegisterService(s *echo.Echo, dbPool *sqlx.DB) *Service {
	service := Service{db: dbPool}
	service.registerRoutes(s)
	postIDGen := generator.NewGenerator()
	service.userCache = NewUserCache()
	service.postIDGenerator = &postIDGen
	return &service
}

func (s *Service) registerRoutes(router *echo.Echo) {
	router.POST("/api/service/clear", s.Clear())
	router.POST("/api/forum/create", s.ForumCreate())
	router.GET("/api/forum/:slug/details", s.ForumGetOne())
	router.GET("/api/forum/:slug/threads", s.ForumGetThreads())
	router.GET("/api/forum/:slug/users", s.ForumGetUsers())
	router.GET("/api/post/:id/details", s.PostGetOne())
	router.POST("/api/post/:id/details", s.UpdatePost())
	router.POST("/api/thread/:slug_or_id/create", s.PostsCreate())
	router.GET("/api/service/status", s.GetStatus())
	router.POST("/api/forum/:slug/create", s.ThreadCreate())
	router.GET("/api/thread/:slug_or_id/details", s.ThreadGetOne())
	router.GET("/api/thread/:slug_or_id/posts", s.ThreadGetPosts())
	router.POST("/api/thread/:slug_or_id/details", s.UpdateThread())
	router.POST("/api/thread/:slug_or_id/vote", s.ThreadVote())
	router.POST("/api/user/:nickname/create", s.UserCreate())
	router.GET("/api/user/:nickname/profile", s.UserGetOne())
	router.POST("/api/user/:nickname/profile", s.UpdateUser())
	return
}

func (s *Service) Clear() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		_, err := s.db.Exec("DELETE FROM votes;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		_, err = s.db.Exec("DELETE FROM posts;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		_, err = s.db.Exec("DELETE FROM thread;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		_, err = s.db.Exec("DELETE FROM forum_users;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		_, err = s.db.Exec("DELETE FROM forum;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		_, err = s.db.Exec("DELETE FROM users;")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		return ctx.NoContent(http.StatusOK)
	}
}

func (s *Service) GetStatus() echo.HandlerFunc {
	return func(ctx echo.Context) error {

		var data Status

		tx, err := s.db.Begin()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}
		defer tx.Rollback()

		err = tx.QueryRow("SELECT count(*) FROM users;").Scan(&data.User)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		err = tx.QueryRow("SELECT count(*) FROM forum;").Scan(&data.Forum)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		err = tx.QueryRow("SELECT count(*) FROM thread;").Scan(&data.Thread)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		err = tx.QueryRow("SELECT count(*) FROM posts;").Scan(&data.Post)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		err = tx.Commit()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}

		return ctx.JSON(http.StatusOK, &data)
	}
}
