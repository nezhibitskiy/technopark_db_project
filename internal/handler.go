package internal

import (
	"encoding/json"
	"errors"
	"github.com/buaazp/fasthttprouter"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"project/internal/consts"
	"project/internal/handlers"
	"project/internal/model"
	"strconv"
	"strings"
)

type Handler struct {
	usecase *Usecase
	router  *fasthttprouter.Router
}

func NewHandler(usecase Usecase, echo *echo.Echo) {
	h := Handler{
		usecase: &usecase,
	}

	echo.POST("/api/user/:nickname/create", h.handleUserCreate())
	echo.GET("/api/user/:nickname/profile", h.handleGetUserProfile())
	echo.POST("/api/user/:nickname/profile", h.handleUserUpdate())
	echo.POST("/api/forum/create", h.handleForumCreate())
	echo.POST("/api/forum/:slug/create", h.handleThreadCreate())
	echo.GET("/api/forum/:slug/details", h.handleGetForumDetails())
	echo.GET("/api/forum/:slug/threads", h.handleGetForumThreads())
	echo.GET("/api/forum/:slug/users", h.handleGetForumUsers())
	echo.POST("/api/thread/:slug_or_id/create", h.handlePostCreate())
	echo.POST("/api/thread/:slug_or_id/vote", h.handleVoteForThread())
	echo.GET("/api/thread/:slug_or_id/details", h.handleGetThreadDetails())
	echo.POST("/api/thread/:slug_or_id/details", h.handleThreadUpdate())
	echo.GET("/api/thread/:slug_or_id/posts", h.handleGetThreadPosts())
	echo.GET("/api/post/:id/details", h.handleGetPostDetails())
	echo.POST("/api/post/:id/details", h.handlePostUpdate())
	echo.GET("/api/service/status", h.handleStatus())
	echo.POST("/api/service/clear", h.handleClear())

}

func (h *Handler) handleUserCreate() echo.HandlerFunc {
	return func(c echo.Context) error {
		u := model.UserInput{}
		requestData, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if err := json.Unmarshal(requestData, &u); err != nil {
			return handlers.BadRequest(c, err)
		}

		users, err := h.usecase.createUser(handlers.PathParam(c, "nickname"), u.Email, u.Fullname, u.About)
		if errors.Is(err, consts.ErrConflict) {
			return handlers.Conflict(c, users)
		}
		if err != nil {
			return handlers.Error(c, err)

		}
		return handlers.Created(c, users[0])
	}
}

func (h *Handler) handleGetUserProfile() echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := h.usecase.getUserByNickname(handlers.PathParam(c, "nickname"))
		if err != nil {
			return handlers.Error(c, err)

		}
		return handlers.Ok(c, u)
	}
}

func (h *Handler) handleUserUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		u := model.UserInput{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return handlers.Error(c, err)
		}
		if err := json.Unmarshal(body, &u); err != nil {
			return handlers.BadRequest(c, err)
		}
		nick := handlers.PathParam(c, "nickname")
		user, err := h.usecase.updateUser(nick, u.Email, u.Fullname, u.About)
		if errors.Is(err, consts.ErrConflict) {
			return handlers.ConflictWithMessage(c, err)

		}
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, user)
	}
}

func (h *Handler) handleForumCreate() echo.HandlerFunc {
	return func(c echo.Context) error {
		forumToCreate := model.ForumCreate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &forumToCreate); err != nil {
			return handlers.BadRequest(c, err)
		}
		forum, err := h.usecase.createForum(forumToCreate.Title, forumToCreate.Slug, forumToCreate.User)
		if errors.Is(err, consts.ErrConflict) {
			return handlers.Conflict(c, forum)

		}
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Created(c, forum)
	}
}

func (h *Handler) handleThreadCreate() echo.HandlerFunc {
	return func(c echo.Context) error {

		thread := model.ThreadCreate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &thread); err != nil {
			return handlers.BadRequest(c, err)
		}
		forum := handlers.PathParam(c, "slug")
		result, err := h.usecase.createThread(forum, thread)
		if errors.Is(err, consts.ErrConflict) {
			return handlers.Conflict(c, result)

		}
		if err != nil {
			return handlers.Error(c, err)

		}
		return handlers.Created(c, result)
	}
}

func (h *Handler) handleGetForumDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		slug := handlers.PathParam(c, "slug")
		forum, err := h.usecase.getForum(slug)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, forum)
	}
}

func (h *Handler) handleGetForumThreads() echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(handlers.QueryParam(c, "limit"))
		desc, _ := strconv.ParseBool(handlers.QueryParam(c, "desc"))
		threads, err := h.usecase.getForumThreads(handlers.PathParam(c, "slug"), handlers.QueryParam(c, "since"), limit, desc)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, threads)
	}
}

func (h *Handler) handleGetForumUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(handlers.QueryParam(c, "limit"))
		desc, _ := strconv.ParseBool(handlers.QueryParam(c, "desc"))
		users, err := h.usecase.getForumUsers(handlers.PathParam(c, "slug"), handlers.QueryParam(c, "since"), limit, desc)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, users)
	}
}

func (h *Handler) handlePostCreate() echo.HandlerFunc {
	return func(c echo.Context) error {

		var posts []*model.PostCreate
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &posts); err != nil {
			return handlers.BadRequest(c, err)
		}
		result, err := h.usecase.createPosts(handlers.PathParam(c, "slug_or_id"), posts)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Created(c, result)
	}
}

func (h *Handler) handleVoteForThread() echo.HandlerFunc {
	return func(c echo.Context) error {

		var vote model.VoteDB
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &vote); err != nil {
			return handlers.BadRequest(c, err)
		}
		thread, err := h.usecase.voteForThread(handlers.PathParam(c, "slug_or_id"), vote)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, thread)
	}
}

func (h *Handler) handleGetThreadDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		thread, err := h.usecase.getThread(handlers.PathParam(c, "slug_or_id"))
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, thread)
	}
}

func (h *Handler) handleThreadUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := model.ThreadUpdate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &t); err != nil {
			return handlers.BadRequest(c, err)
		}
		thread, err := h.usecase.updateThread(handlers.PathParam(c, "slug_or_id"), t.Message, t.Title)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, thread)
	}
}

func (h *Handler) handleGetThreadPosts() echo.HandlerFunc {
	return func(c echo.Context) error {
		sp := handlers.QueryParam(c, "since")
		var since *int = nil
		if sp != "" {
			n, _ := strconv.Atoi(sp)
			since = &n
		}
		limit, _ := strconv.Atoi(handlers.QueryParam(c, "limit"))
		desc, _ := strconv.ParseBool(handlers.QueryParam(c, "desc"))
		posts, err := h.usecase.getThreadPosts(
			handlers.PathParam(c, "slug_or_id"),
			limit,
			since,
			handlers.QueryParam(c, "sort"),
			desc,
		)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, posts)
	}
}

func (h *Handler) handleGetPostDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(handlers.PathParam(c, "id"))
		related := strings.Split(handlers.QueryParam(c, "related"), ",")
		details, err := h.usecase.getPostDetails(id, related)
		if err != nil {
			return handlers.Error(c, err)
		}
		result := map[string]interface{}{
			"post": details.Post,
		}
		for _, r := range related {
			switch r {
			case "user":
				result["author"] = details.Author
			case "forum":
				result["forum"] = details.Forum
			case "thread":
				result["thread"] = details.Thread
			}
		}
		return handlers.Ok(c, result)
	}
}

func (h *Handler) handlePostUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := model.PostUpdate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &t); err != nil {
			return handlers.BadRequest(c, err)
		}
		id, _ := strconv.Atoi(handlers.PathParam(c, "id"))
		thread, err := h.usecase.updatePost(id, t.Message)
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, thread)
	}
}

func (h *Handler) handleStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		status, err := h.usecase.getStatus()
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, status)
	}
}

func (h *Handler) handleClear() echo.HandlerFunc {
	return func(c echo.Context) error {
		err := h.usecase.clear()
		if err != nil {
			return handlers.Error(c, err)
		}
		return handlers.Ok(c, nil)
	}
}
