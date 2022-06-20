package internal

import (
	"encoding/json"
	"errors"
	"github.com/buaazp/fasthttprouter"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"project/internal/consts"
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
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}

		users, err := h.usecase.createUser(c.Param("nickname"), u.Email, u.Fullname, u.About)
		if errors.Is(err, consts.ErrConflict) {
			return c.JSON(http.StatusConflict, &users)
		}
		if err != nil {
			return Error(c, err)

		}
		return c.JSON(http.StatusCreated, users[0])
	}
}

func (h *Handler) handleGetUserProfile() echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := h.usecase.getUserByNickname(c.Param("nickname"))
		if err != nil {
			return Error(c, err)

		}
		return c.JSON(http.StatusOK, u)
	}
}

func (h *Handler) handleUserUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		u := model.UserInput{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return Error(c, err)
		}
		if err := json.Unmarshal(body, &u); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		nick := c.Param("nickname")
		user, err := h.usecase.updateUser(nick, u.Email, u.Fullname, u.About)
		if errors.Is(err, consts.ErrConflict) {
			return c.JSON(http.StatusConflict, map[string]string{
				"message": err.Error(),
			})
		}
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, user)
	}
}

func (h *Handler) handleForumCreate() echo.HandlerFunc {
	return func(c echo.Context) error {
		forumToCreate := model.ForumCreate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &forumToCreate); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		forum, err := h.usecase.createForum(forumToCreate.Title, forumToCreate.Slug, forumToCreate.User)
		if errors.Is(err, consts.ErrConflict) {
			return c.JSON(http.StatusConflict, &forum)

		}
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusCreated, forum)
	}
}

func (h *Handler) handleThreadCreate() echo.HandlerFunc {
	return func(c echo.Context) error {

		thread := model.ThreadCreate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &thread); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		forum := c.Param("slug")
		result, err := h.usecase.createThread(forum, thread)
		if errors.Is(err, consts.ErrConflict) {
			return c.JSON(http.StatusConflict, &result)

		}
		if err != nil {
			return Error(c, err)

		}
		return c.JSON(http.StatusCreated, result)
	}
}

func (h *Handler) handleGetForumDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		slug := c.Param("slug")
		forum, err := h.usecase.getForum(slug)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, forum)
	}
}

func (h *Handler) handleGetForumThreads() echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		desc, _ := strconv.ParseBool(c.QueryParam("desc"))
		threads, err := h.usecase.getForumThreads(c.Param("slug"), c.QueryParam("since"), limit, desc)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, threads)
	}
}

func (h *Handler) handleGetForumUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		desc, _ := strconv.ParseBool(c.QueryParam("desc"))
		users, err := h.usecase.getForumUsers(c.Param("slug"), c.QueryParam("since"), limit, desc)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, users)
	}
}

func (h *Handler) handlePostCreate() echo.HandlerFunc {
	return func(c echo.Context) error {

		var posts []*model.PostCreate
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &posts); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		result, err := h.usecase.createPosts(c.Param("slug_or_id"), posts)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusCreated, result)
	}
}

func (h *Handler) handleVoteForThread() echo.HandlerFunc {
	return func(c echo.Context) error {

		var vote model.VoteDB
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &vote); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		thread, err := h.usecase.voteForThread(c.Param("slug_or_id"), vote)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, thread)
	}
}

func (h *Handler) handleGetThreadDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		thread, err := h.usecase.getThread(c.Param("slug_or_id"))
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, thread)
	}
}

func (h *Handler) handleThreadUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := model.ThreadUpdate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &t); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		thread, err := h.usecase.updateThread(c.Param("slug_or_id"), t.Message, t.Title)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, thread)
	}
}

func (h *Handler) handleGetThreadPosts() echo.HandlerFunc {
	return func(c echo.Context) error {
		sp := c.QueryParam("since")
		var since *int = nil
		if sp != "" {
			n, _ := strconv.Atoi(sp)
			since = &n
		}
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		desc, _ := strconv.ParseBool(c.QueryParam("desc"))
		posts, err := h.usecase.getThreadPosts(
			c.Param("slug_or_id"),
			limit,
			since,
			c.QueryParam("sort"),
			desc,
		)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, posts)
	}
}

func (h *Handler) handleGetPostDetails() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		related := strings.Split(c.QueryParam("related"), ",")
		details, err := h.usecase.getPostDetails(id, related)
		if err != nil {
			return Error(c, err)
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
		return c.JSON(http.StatusOK, result)
	}
}

func (h *Handler) handlePostUpdate() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := model.PostUpdate{}
		body, err := ioutil.ReadAll(c.Request().Body)
		if err := json.Unmarshal(body, &t); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
		}
		id, _ := strconv.Atoi(c.Param("id"))
		thread, err := h.usecase.updatePost(id, t.Message)
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, thread)
	}
}

func (h *Handler) handleStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		status, err := h.usecase.getStatus()
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, status)
	}
}

func (h *Handler) handleClear() echo.HandlerFunc {
	return func(c echo.Context) error {
		err := h.usecase.clear()
		if err != nil {
			return Error(c, err)
		}
		return c.JSON(http.StatusOK, nil)
	}
}

func Error(c echo.Context, err error) error {
	if errors.Is(err, consts.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": err.Error(),
		})
	}
	if errors.Is(err, consts.ErrConflict) {
		return c.JSON(http.StatusConflict, map[string]string{
			"message": err.Error(),
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	return c.NoContent(http.StatusInternalServerError)
}
