package internal

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	pathDelim    = "."
	maxIDLength  = 7
	maxTreeLevel = 5
)

var zeroPathStud = strings.Repeat("0", maxIDLength)

func (s *Service) PostGetOne() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Post{}

		queryParam := ctx.Param("id")
		relatedParam := ctx.QueryParam("related")

		id, err := strconv.Atoi(queryParam)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}

		data.Id = uint32(id)
		err = s.db.QueryRow(context.Background(), "SELECT parent, author, message, is_edited, thread_id, "+
			"created_at FROM posts WHERE id = $1;", data.Id).Scan(&data.Parent, &data.Author, &data.Message,
			&data.IsEdited, &data.Thread, &data.Created)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}

		thread := Thread{}
		err = s.db.QueryRow(context.Background(), "SELECT id, slug, title, author, forum, message, "+
			"created_at FROM thread WHERE id = $1", &data.Thread).Scan(&thread.Id, &thread.Slug, &thread.Title,
			&thread.Author, &thread.Forum, &thread.Message, &thread.Created)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}
		data.Forum = thread.Forum

		postFull := PostFull{Post: &data}

		relatedParams := strings.Split(relatedParam, ",")

		for _, relParam := range relatedParams {
			switch relParam {
			case "user":
				user, err := s.userCache.GetUserByNickname(data.Author)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user by nickname: " + data.Author})
				}
				postFull.Author = user

			case "forum":
				forum := Forum{}
				conn, err := s.db.Acquire(context.Background())
				defer conn.Release()

				tx, err := conn.Begin(context.Background())
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
				}
				defer tx.Rollback(context.Background())

				err = tx.QueryRow(context.Background(), "SELECT title, author, slug FROM forum WHERE slug=$1;", &data.Forum).Scan(&forum.Title, &forum.User, &forum.Slug)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}

				err = tx.QueryRow(context.Background(), "SELECT count(*) FROM thread WHERE forum=$1;", &data.Forum).Scan(&forum.Threads)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}

				err = tx.QueryRow(context.Background(), "SELECT count(*) FROM posts RIGHT JOIN thread t on t.id = posts.thread_id WHERE t.forum=$1;", &data.Forum).Scan(&forum.Posts)
				if err != nil {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}

				err = tx.Commit(context.Background())
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
				}
				postFull.Forum = &forum

			case "thread":
				postFull.Thread = &thread
			}
		}

		return ctx.JSON(http.StatusOK, &postFull)
	}
}

func (s *Service) UpdatePost() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		postUPD := PostUpdate{}
		queryParam := ctx.Param("id")
		err := ctx.Bind(&postUPD)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		data := Post{}
		err = s.db.QueryRow(context.Background(), "SELECT id, parent, author, message, is_edited, thread_id, "+
			"created_at FROM posts WHERE id = $1;", &queryParam).Scan(&data.Id, &data.Parent, &data.Author, &data.Message,
			&data.IsEdited, &data.Thread, &data.Created)
		err = s.db.QueryRow(context.Background(), "SELECT forum FROM thread WHERE id = $1;",
			data.Thread).Scan(&data.Forum)

		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}

		if postUPD.Message != data.Message && postUPD.Message != "" {
			_, err = s.db.Exec(context.Background(), "UPDATE posts SET message = $1, is_edited = true WHERE id = $2",
				&postUPD.Message, &queryParam)
			data.Message = postUPD.Message
			data.IsEdited = true
			if err != nil {
				if err == pgx.ErrNoRows {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
				}
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}
		}

		return ctx.JSON(http.StatusOK, &data)
	}
}

func (s *Service) PostsCreate() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		postsArr := make([]Post, 0, 8)
		created := time.Now()
		err := ctx.Bind(&postsArr)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		slugOrId := ctx.Param("slug_or_id")

		forum, id, err := s.GetForumAndIDFromThread(slugOrId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: err.Error()})
		}

		for i, _ := range postsArr {
			//var thisTime time.Time
			//if !postsArr[0].Created.IsZero() {
			//	thisTime = postsArr[0].Created.UTC()
			//}

			if postsArr[i].Thread == 0 {
				postsArr[i].Thread = id
			}

			if postsArr[i].Parent != 0 {
				parentThread := 0
				err = s.db.QueryRow(context.Background(), "SELECT thread_id FROM posts WHERE id = $1", postsArr[i].Parent).Scan(&parentThread)
				if parentThread != postsArr[i].Thread {
					return ctx.JSON(http.StatusConflict, ResponseError{Message: "Parent post was created in another thread"})
				}
			}

			postsArr[i].Id = s.postIDGenerator.Next()
			postsArr[i].Path, err = s.getPostPath(postsArr[i].Id, postsArr[i].Parent)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}

			err = s.db.QueryRow(context.Background(), "INSERT INTO posts(id, author, path, parent, message, thread_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
				&postsArr[i].Id, &postsArr[i].Author, &postsArr[i].Path, &postsArr[i].Parent, &postsArr[i].Message, &postsArr[i].Thread, &created).Scan(&postsArr[i].Id)
			if err != nil {
				if strings.Contains(err.Error(), "23503") {
					return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find post author by nickname: " + postsArr[i].Author})
				}
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}

			postsArr[i].Forum = forum
			_, err = s.db.Exec(context.Background(), "INSERT INTO forum_users(author, forum) VALUES ($1, $2);", &postsArr[i].Author, &postsArr[i].Forum)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, ResponseError{Message: err.Error()})
			}
		}

		return ctx.JSON(http.StatusCreated, &postsArr)
	}
}

func (s *Service) getPostPath(id uint32, parentID uint) (string, error) {
	var base string
	if parentID == 0 {
		base = s.getZeroPostPath()
	} else {
		parentPath := ""
		err := s.db.QueryRow(context.Background(), "SELECT path FROM posts WHERE id = $1", &parentID).Scan(&parentPath)
		if err != nil {
			return "", err
		}
		base = parentPath
	}
	path := strings.Replace(base, zeroPathStud, s.padPostID(id), 1)
	return path, nil
}

func (s *Service) padPostID(id uint32) string {
	return fmt.Sprintf("%0"+strconv.Itoa(maxIDLength)+"d", id)
}

func (s *Service) getZeroPostPath() string {
	path := zeroPathStud
	for i := 0; i < maxTreeLevel-1; i++ {
		path += pathDelim + zeroPathStud
	}
	return path
}
