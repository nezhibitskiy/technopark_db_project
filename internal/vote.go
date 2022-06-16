package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

var errVoteExists = errors.New("vote is exists")

func (s *Service) FindUserVote(data Vote, threadId uint) (int, error) {
	vote := 0
	value := 0
	err := s.db.QueryRow("SELECT id, value FROM votes WHERE author = $1 AND thread_id = $2",
		&data.Nickname, threadId).Scan(&vote, &value)
	if err != nil {
		return 0, err
	}
	if value == data.Voice {
		return vote, errVoteExists
	}
	return vote, nil
}

func (s *Service) ThreadVote() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Vote{}

		slug := ctx.Param("slug_or_id")
		err := json.NewDecoder(ctx.Request().Body).Decode(&data)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err)
		}
		user, err := s.userCache.GetUserByNickname(data.Nickname)
		if user == nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user with nickname: " + data.Nickname})
		}

		var thread *Thread
		id, err := strconv.Atoi(slug)
		if err != nil {
			thread, err = s.GetThreadByIDorSlug(slug, 0)
			if err != nil && err != sql.ErrNoRows {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			if thread == nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by slug: " + slug})
			}
		} else {
			thread, err = s.GetThreadByIDorSlug("", uint(id))
			if err != nil && err != sql.ErrNoRows {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			if thread == nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by id: " + slug})
			}

		}

		voteID, err := s.FindUserVote(data, thread.Id)
		if err != nil && err != sql.ErrNoRows {
			if err == errVoteExists {
				err = s.FillThreadVotes(thread)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, err)
				}
				return ctx.JSON(http.StatusOK, &thread)
			}
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if err == sql.ErrNoRows {

			_, err = s.db.Exec("INSERT INTO votes(thread_id, author, value) VALUES($1, $2, $3);",
				&thread.Id, &data.Nickname, &data.Voice)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}

			err = s.db.QueryRow("UPDATE thread SET votes = votes + $1 WHERE id = $2 "+
				"RETURNING votes", &data.Voice, &thread.Id).Scan(&thread.Votes)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		} else {
			_, err = s.db.Exec("UPDATE votes SET value = $1 WHERE id = $2;",
				&data.Voice, &voteID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			data.Voice = data.Voice * 2
			err = s.db.QueryRow("UPDATE thread SET votes = votes + $1 WHERE id = $2 "+
				"RETURNING votes", &data.Voice, &thread.Id).Scan(&thread.Votes)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		}
		return ctx.JSON(http.StatusOK, &thread)
	}
}
