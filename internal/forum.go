package internal

import (
	"database/sql"
	"encoding/json"
	"github.com/labstack/echo/v4"

	"net/http"
)

func (s *Service) getForumBySlug(slug string) (*Forum, error) {
	data := Forum{}
	err := s.db.QueryRow("SELECT title, author, slug FROM forum WHERE slug=$1;", &slug).Scan(&data.Title, &data.User, &data.Slug)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Service) ForumCreate() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Forum{}
		err := json.NewDecoder(ctx.Request().Body).Decode(&data)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err)
		}

		user, err := s.searchUsersByNickname(data.User)
		if err != nil && err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if user == nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user with nickname: " + data.User})
		}
		data.User = user.Nickname

		oldForum, err := s.getForumBySlug(data.Slug)
		if err != nil && err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if oldForum != nil {
			return ctx.JSON(http.StatusConflict, oldForum)
		}

		_, err = s.db.Exec("INSERT INTO forum(title, slug, author) VALUES($1, $2, $3);",
			&data.Title, &data.Slug, &data.User)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		return ctx.JSON(http.StatusCreated, &data)
	}
}

func (s *Service) ForumGetOne() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		forumSlug := ctx.Param("slug")
		data := Forum{}

		err := s.db.QueryRow("SELECT title, author, slug, posts, threads FROM forum "+
			"WHERE slug=$1;", &forumSlug).Scan(&data.Title, &data.User, &data.Slug, &data.Posts, &data.Threads)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}

		return ctx.JSON(http.StatusOK, &data)
	}
}

func (s *Service) ForumGetThreads() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		forumSlug := ctx.Param("slug")
		threads := make([]Thread, 0, 8)

		limitStr := ctx.QueryParam("limit")
		descStr := ctx.QueryParam("desc")
		sinceStr := ctx.QueryParam("since")

		forumTitle := ""
		err := s.db.QueryRow("SELECT title FROM forum WHERE slug = $1", &forumSlug).Scan(&forumTitle)
		if err != nil && err != sql.ErrNoRows {
			return ctx.NoContent(http.StatusInternalServerError)
		}
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find forum by slug: " + forumSlug})
		}

		sql := "SELECT id, title, author, forum, message, slug, created_at, votes FROM thread WHERE forum = $1"

		if sinceStr != "" {
			if descStr == "true" {
				sql = sql + " AND created_at <= '" + sinceStr + "'"
			} else {
				sql = sql + " AND created_at >= '" + sinceStr + "'"
			}
		}
		sql = sql + " ORDER BY created_at"
		if descStr == "true" {
			sql = sql + " DESC"
		}
		if limitStr != "" {
			sql = sql + " LIMIT " + limitStr
		}
		sql = sql + ";"

		err = s.db.Select(&threads, sql, forumSlug)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}
		return ctx.JSON(http.StatusOK, &threads)
	}
}

func (s *Service) ForumGetUsers() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		forumSlug := ctx.Param("slug")
		descStr := ctx.QueryParam("desc")
		limitStr := ctx.QueryParam("limit")
		sinceStr := ctx.QueryParam("since")
		users := make([]User, 0, 8)

		sql := "SELECT slug FROM forum WHERE slug = $1"
		err := s.db.QueryRow(sql, forumSlug).Scan(&forumSlug)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find forum by slug: " + forumSlug})
		}

		sql = "SELECT DISTINCT nickname, fullname, about, email FROM users " +
			"JOIN forum_users fu on users.nickname = fu.author WHERE fu.forum = $1"

		if sinceStr != "" {
			if descStr == "true" {
				sql = sql + " AND nickname < '" + sinceStr + "'"
			} else {
				sql = sql + " AND nickname > '" + sinceStr + "'"
			}
		}

		sql = sql + " ORDER BY nickname"

		if descStr == "true" {
			sql = sql + " DESC"
		}

		if limitStr != "" {
			sql = sql + " LIMIT " + limitStr
		}

		err = s.db.Select(&users, sql, forumSlug)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
		}
		return ctx.JSON(http.StatusOK, &users)
	}
}
