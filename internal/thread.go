package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

const ErrNoThreadByID = "Can't find post thread by id: "
const ErrNoThreadBySlug = "Can't find post thread by slug: "

func (s *Service) GetForumAndIDFromThread(slugOrId string) (string, int, error) {
	forum := ""
	id := 0
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		err = s.db.QueryRow("SELECT forum, id FROM thread WHERE slug = $1",
			&slugOrId).Scan(&forum, &id)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", 0, errors.New(ErrNoThreadBySlug + slugOrId)
			}
			return "", 0, err
		}
	} else {
		err = s.db.QueryRow("SELECT forum, id FROM thread WHERE id = $1",
			&id).Scan(&forum, &id)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", 0, errors.New(ErrNoThreadByID + slugOrId)
			}
			return "", 0, err
		}
	}
	return forum, id, nil
}

func (s *Service) FillThreadVotes(data *Thread) error {
	return s.db.QueryRow("SELECT votes FROM thread WHERE id = $1", &data.Id).Scan(&data.Votes)
}

func (s *Service) GetThreadByIDorSlug(Slug string, Id uint) (*Thread, error) {
	data := Thread{}
	if Slug != "" {
		err := s.db.QueryRow("SELECT id, slug, title, author, forum, message, created_at, votes "+
			"FROM thread WHERE slug = $1;", &Slug).Scan(&data.Id, &data.Slug, &data.Title, &data.Author, &data.Forum,
			&data.Message, &data.CreatedAt, &data.Votes)
		if err != nil {
			return nil, err
		}
		return &data, nil
	} else if Id != 0 {
		err := s.db.QueryRow("SELECT id, slug, title, author, forum, message, created_at, votes "+
			"FROM thread WHERE id = $1;", &Id).Scan(&data.Id, &data.Slug, &data.Title, &data.Author, &data.Forum,
			&data.Message, &data.CreatedAt, &data.Votes)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
	return nil, nil
}

func (s *Service) IsAuthorExists(nickname string) (bool, error) {
	fullname := ""
	err := s.db.QueryRow("SELECT fullname FROM users WHERE nickname = $1;", &nickname).Scan(&fullname)
	if err != nil {
		return false, err
	}
	if fullname != "" {
		return true, nil
	}
	return false, nil
}

func (s *Service) ThreadCreate() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Thread{}
		err := json.NewDecoder(ctx.Request().Body).Decode(&data)
		slug := ctx.Param("slug")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		oldThread, err := s.GetThreadByIDorSlug(data.Slug, data.Id)
		if err != nil && err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		authorExists, err := s.IsAuthorExists(data.Author)
		if err != nil && err != sql.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		if !authorExists {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread author by nickname: " + data.Author})
		}

		if oldThread == nil {

			tx, err := s.db.Begin()
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}
			defer tx.Rollback()

			err = tx.QueryRow("SELECT slug FROM forum WHERE slug = $1", slug).Scan(&data.Forum)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread forum by slug: " + slug})
			}

			err = tx.QueryRow("INSERT INTO thread(slug, title, message, author, forum, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
				&data.Slug, &data.Title, &data.Message, &data.Author, &data.Forum, &data.CreatedAt).Scan(&data.Id)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}

			err = tx.Commit()
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}

			return ctx.JSON(http.StatusCreated, &data)
		}
		return ctx.JSON(http.StatusConflict, &oldThread)
	}
}
func (s *Service) ThreadGetOne() echo.HandlerFunc {
	return func(ctx echo.Context) error {

		var data *Thread
		queryParam := ctx.Param("slug_or_id")

		id, err := strconv.Atoi(queryParam)
		if err != nil {
			data, err = s.GetThreadByIDorSlug(queryParam, 0)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
			}
		} else {
			data, err = s.GetThreadByIDorSlug("", uint(id))
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
			}
		}
		return ctx.JSON(http.StatusOK, &data)
	}
}
func (s *Service) ThreadGetPosts() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sortParam := ctx.QueryParam("sort")
		limitParam := ctx.QueryParam("limit")
		slugOrId := ctx.Param("slug_or_id")
		sinceParam := ctx.QueryParam("since")

		var since *int
		if sinceParam != "" {
			n, _ := strconv.Atoi(sinceParam)
			since = &n
		}

		orderParam := ctx.QueryParam("desc")
		desc := false
		if orderParam == "true" {
			desc = true
		}
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		id, err := strconv.Atoi(slugOrId)
		if err != nil {
			err = s.db.QueryRow("SELECT id FROM thread WHERE slug = $1",
				&slugOrId).Scan(&id)
			if err != nil && err == sql.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find post thread by slug: " + slugOrId})
			}
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}
		} else {
			err = s.db.QueryRow("SELECT id FROM thread WHERE id = $1",
				&slugOrId).Scan(&id)
			if err != nil && err == sql.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find post thread by id: " + slugOrId})
			}
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}
		}

		posts, err := s.GetThreadPosts(id, limit, since, sortParam, desc)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		return ctx.JSON(http.StatusOK, posts)
		//switch sortParam {
		//case "flat":
		//	posts, err := s.ThreadPostsFlat(id, limit, since, desc)
		//	if err != nil {
		//		return ctx.JSON(http.StatusInternalServerError, err)
		//	}
		//	return ctx.JSON(http.StatusOK, posts)
		//
		//case "tree":
		//	posts, err := s.ThreadPostsTree(id, limit, since, desc)
		//	if err != nil {
		//		return ctx.JSON(http.StatusInternalServerError, err)
		//	}
		//	return ctx.JSON(http.StatusOK, posts)
		//case "parent_tree":
		//	posts, err := s.ThreadPostsParentTree(id, limit, since, desc)
		//	if err != nil {
		//		return ctx.JSON(http.StatusInternalServerError, err)
		//	}
		//	return ctx.JSON(http.StatusOK, posts)
		//default:
		//	posts, err := s.ThreadPostsFlat(id, limit, since, desc)
		//	if err != nil {
		//		return ctx.JSON(http.StatusInternalServerError, err)
		//	}
		//	return ctx.JSON(http.StatusOK, posts)
		//
		//}
	}
}
func (s *Service) UpdateThread() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Thread{}
		err := json.NewDecoder(ctx.Request().Body).Decode(&data)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		queryParam := ctx.Param("slug_or_id")
		id, err := strconv.Atoi(queryParam)
		if err != nil {
			err = s.db.QueryRow("SELECT slug FROM thread WHERE slug = $1;",
				&queryParam).Scan(&data.Slug)
			if err != nil {
				if err == sql.ErrNoRows {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by slug: " + queryParam})
				}
			}
			if data.Message != "" && data.Title != "" {
				data.Slug = queryParam
				_, err = s.db.Exec("UPDATE thread SET title=$1, message=$2 WHERE slug = $3;", &data.Title, &data.Message, &queryParam)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}
			} else if data.Message != "" {
				data.Slug = queryParam
				_, err = s.db.Exec("UPDATE thread SET message=$1 WHERE slug = $2;", &data.Message, &queryParam)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}
			} else if data.Title != "" {
				data.Slug = queryParam
				_, err = s.db.Exec("UPDATE thread SET title=$1 WHERE slug = $2;", &data.Title, &queryParam)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}
			}

			err = s.db.QueryRow("SELECT id, slug, title, author, forum, message, created_at "+
				"FROM thread WHERE slug = $1;", &queryParam).Scan(&data.Id, &data.Slug, &data.Title, &data.Author, &data.Forum,
				&data.Message, &data.CreatedAt)
			return ctx.JSON(http.StatusOK, &data)
		}

		err = s.db.QueryRow("SELECT id FROM thread WHERE id = $1;",
			&queryParam).Scan(&data.Id)
		if err != nil {
			if err == sql.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by id: " + queryParam})
			}
		}

		if data.Message != "" && data.Title != "" {
			data.Id = uint(id)
			_, err = s.db.Exec("UPDATE thread SET title=$1, message=$2 WHERE id = $3;", &data.Title, &data.Message, &id)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
			}
		} else if data.Message != "" {
			data.Slug = queryParam
			_, err = s.db.Exec("UPDATE thread SET message=$1 WHERE id = $2;", &data.Message, &id)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
			}
		} else if data.Title != "" {
			data.Slug = queryParam
			_, err = s.db.Exec("UPDATE thread SET title=$1 WHERE id = $2;", &data.Title, &id)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
			}
		}

		err = s.db.QueryRow("SELECT title, author, forum, message, slug, created_at "+
			"FROM thread WHERE id = $1;", data.Id).Scan(&data.Title, &data.Author, &data.Forum,
			&data.Message, &data.Slug, &data.CreatedAt)

		return ctx.JSON(http.StatusOK, &data)
	}
}

func (s *Service) ThreadPostsFlat(thread, limit int, since *int, desc bool) ([]Post, error) {

	sql := "SELECT id, parent, author, message, is_edited, thread, created_at, forum FROM posts " +
		"WHERE thread = $1 "
	if since != nil {
		if desc {
			sql = fmt.Sprintf("%s %s %d ", sql, "AND id <", *since)
		} else {
			sql = fmt.Sprintf("%s %s %d ", sql, "AND id >", *since)
		}
	}
	if desc {
		sql = sql + "ORDER BY created_at DESC, id DESC"
	} else {
		sql = sql + "ORDER BY created_at, id"
	}
	if limit != 0 {
		sql = fmt.Sprintf("%s %s %d", sql, "LIMIT", limit)
	}

	posts := make([]Post, 0, 8)

	err := s.db.Select(&posts, sql, &thread)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *Service) ThreadPostsTree(thread, limit int, since *int, desc bool) ([]Post, error) {
	sql := "SELECT id, parent, author, message, is_edited, thread, created_at FROM posts " +
		"WHERE thread = $1 "
	if since != nil {
		sinceCondition, err := s.getSinceCondition(since, desc)
		if err != nil {
			return nil, err
		}
		sql = fmt.Sprintf("%sAND %s ", sql, sinceCondition)
	}
	if desc {
		sql = sql + "ORDER BY path DESC"
	} else {
		sql = sql + "ORDER BY path"
	}
	if limit != 0 {
		sql = fmt.Sprintf("%s %s %d", sql, "LIMIT", limit)
	}

	forum := ""
	err := s.db.QueryRow("SELECT forum FROM thread WHERE id = $1", &thread).Scan(&forum)
	if err != nil {
		return nil, err
	}

	posts := make([]Post, 0, 8)
	err = s.db.Select(&posts, sql, &thread)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *Service) ThreadPostsParentTree(thread, limit int, since *int, desc bool) ([]Post, error) {

	sql := "SELECT id, parent, author, message, is_edited, thread, created_at FROM posts WHERE thread = $1 "
	if since != nil {
		var operator = ">"
		if desc {
			operator = "<"
		}
		sincePost := ""
		err := s.db.QueryRow("SELECT path FROM posts WHERE id = $1", since).Scan(&sincePost)
		if err != nil {
			return nil, err
		}
		sinceCond := fmt.Sprintf("path %s '%s'", operator, s.getRootPath(sincePost))

		sql = fmt.Sprintf("%sAND %s ", sql, sinceCond)
	}
	sql = sql + " AND parent = 0 ORDER BY id"
	if desc {
		sql = sql + " DESC"
	}

	if limit != 0 {
		sql = fmt.Sprintf(" %s %s %d", sql, "LIMIT", limit)
	}

	forum := ""
	err := s.db.QueryRow("SELECT forum FROM thread WHERE id = $1", &thread).Scan(&forum)
	if err != nil {
		return nil, err
	}

	parents := make([]Post, 0, 8)
	rows, err := s.db.Query(sql, &thread)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		post := Post{}
		post.Forum = forum
		err = rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Thread, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		parents = append(parents, post)
	}

	posts := make([]Post, 0, 8)
	for _, parent := range parents {
		var childs []Post

		sql = fmt.Sprintf(`SELECT id, parent, author, message, is_edited, thread, created_at FROM posts WHERE substring(path,1,7) = '%s' AND parent<>0 ORDER BY path`, s.padPostID(parent.Id))

		rows, err = s.db.Query(sql)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			post := Post{}
			post.Forum = forum
			err = rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Thread, &post.CreatedAt)
			if err != nil {
				return nil, err
			}
			childs = append(childs, post)
		}
		posts = append(posts, parent)
		posts = append(posts, childs...)
	}
	return posts, nil
}

func (s *Service) getRootPath(path string) string {
	root := strings.Split(path, pathDelim)[0]
	return root + strings.Repeat(pathDelim+zeroPathStud, maxTreeLevel-1)
}

func (s *Service) getSinceCondition(since *int, desc bool) (string, error) {
	var operator = ">"
	if desc {
		operator = "<"
	}
	sincePost := ""
	err := s.db.QueryRow("SELECT path FROM posts WHERE id = $1", since).Scan(&sincePost)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("path %s '%s'", operator, sincePost), nil
}

func (s *Service) GetThreadPosts(thread, limit int, since *int, sort string, desc bool) ([]Post, error) {
	switch sort {
	case "flat":
		return s.getThreadPostsFlat(thread, limit, since, desc)
	case "tree":
		return s.getThreadPostsTree(thread, limit, since, desc)
	case "parent_tree":
		return s.getThreadPostsParentTree(thread, limit, since, desc)
	}
	return nil, fmt.Errorf("unknown sort method")
}

func (s *Service) getThreadPostsFlat(thread, limit int, since *int, desc bool) ([]Post, error) {
	order := "asc"
	if desc {
		order = "desc"
	}
	orderBy := []string{"created_at " + order, "id " + order}
	filter := "thread = $1"
	params := []interface{}{thread}
	if since != nil {
		if desc {
			filter += " and id < $2"
		} else {
			filter += " and id > $2"
		}
		params = append(params, *since)
	}
	return s.getPosts(orderBy, limit, filter, params...)
}

func (s *Service) getThreadPostsTree(thread, limit int, since *int, desc bool) ([]Post, error) {
	conditions := []string{"thread = $1"}
	params := []interface{}{thread}
	if since != nil {
		sinceCond, err := s.getSinceCondition(since, desc)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, sinceCond)
	}

	orderBy := []string{"path " + s.getOrder(desc)}
	filter := strings.Join(conditions, " and ")
	return s.getPosts(orderBy, limit, filter, params...)
}

func (s *Service) getThreadPostsParentTree(thread, limit int, since *int, desc bool) ([]Post, error) {
	conditions := []string{"parent=0", "thread=$1"}

	if since != nil {
		var operator = ">"
		if desc {
			operator = "<"
		}
		sincePost, err := s.getPostFields("path", "id=$1", *since)
		if err != nil {
			return nil, err
		}
		sinceCond := fmt.Sprintf("path %s '%s'", operator, s.getRootPath(sincePost.Path))
		conditions = append(conditions, sinceCond)
	}

	filter := strings.Join(conditions, " and ")
	var parents []Post
	err := s.db.Select(&parents, fmt.Sprintf(
		`select * from posts where %s order by id %s limit %d`, filter, s.getOrder(desc), limit),
		thread,
	)
	if err != nil {
		return nil, err
	}
	posts := make([]Post, 0)
	for _, parent := range parents {
		var childs []Post
		err := s.db.Select(&childs, fmt.Sprintf(
			`select * from posts where substring(path,1,7) = '%s' and parent<>0 order by path`, s.padPostID(parent.Id),
		))
		if err != nil {
			return nil, err
		}
		posts = append(posts, parent)
		posts = append(posts, childs...)
	}
	return posts, nil
}

func (s *Service) getOrder(desc bool) string {
	if desc {
		return " desc"
	}
	return ""
}

func (s *Service) getPostFields(fields, filter string, params ...interface{}) (*Post, error) {
	p := Post{}
	err := s.db.Get(&p, "select "+fields+" from posts where "+filter, params...)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
