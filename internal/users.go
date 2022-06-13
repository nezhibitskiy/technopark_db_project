package internal

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *Service) searchUsersByNickname(nickname string) (*User, error) {
	user := User{}

	err := s.db.QueryRow(context.Background(), "SELECT nickname, fullname, about, email FROM users "+
		"WHERE nickname = $1", &nickname).Scan(&user.Nickname, &user.Fullname, &user.About,
		&user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) searchUsersByEmail(email string) ([]User, error) {
	users := make([]User, 0, 2)

	rows, err := s.db.Query(context.Background(), "SELECT nickname, fullname, about, email FROM users "+
		"WHERE email = $1", &email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nextUser User
		err = rows.Scan(&nextUser.Nickname, &nextUser.Fullname, &nextUser.About,
			&nextUser.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, nextUser)
	}

	return users, nil
}

func (s *Service) searchUsersByEmailOrNickname(nickname, email string) ([]User, error) {
	users := make([]User, 0, 2)

	rows, err := s.db.Query(context.Background(), "SELECT nickname, fullname, about, email FROM users "+
		"WHERE nickname = $1 OR email = $2", &nickname, &email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var nextUser User
		err = rows.Scan(&nextUser.Nickname, &nextUser.Fullname, &nextUser.About,
			&nextUser.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, nextUser)
	}

	return users, nil
}

func (s *Service) getUserByNickname(nickname string) (User, error) {
	userData := User{}
	err := s.db.QueryRow(context.Background(), "SELECT nickname, fullname, about, email FROM users "+
		"WHERE nickname = $1", &nickname).Scan(&userData.Nickname, &userData.Fullname, &userData.About,
		&userData.Email)
	return userData, err
}

func (s *Service) UserCreate() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		nickname := ctx.Param("nickname")
		userData := User{}
		err := ctx.Bind(&userData)
		if err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		oldUsers, err := s.searchUsersByEmailOrNickname(nickname, userData.Email)
		if len(oldUsers) != 0 {
			return ctx.JSON(http.StatusConflict, oldUsers)
		}

		_, err = s.db.Exec(context.Background(), "INSERT INTO users(about, email, fullname, nickname) "+
			"VALUES($1, $2, $3, $4);", &userData.About, &userData.Email,
			&userData.Fullname, &nickname)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		userData.Nickname = nickname
		s.userCache.Add(&userData)
		return ctx.JSON(http.StatusCreated, &userData)
	}
}

func (s *Service) UserGetOne() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		nickname := ctx.Param("nickname")
		userData, err := s.userCache.GetUserByNickname(nickname)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user by nickname: " + nickname})
		}
		return ctx.JSON(http.StatusOK, &userData)
	}
}

func (s *Service) UpdateUser() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		nickname := ctx.Param("nickname")
		userData := User{}
		err := ctx.Bind(&userData)
		if err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		var oldUsersEmail *User

		oldUsersEmail, err = s.userCache.GetUserByEmail(userData.Email)
		if err != nil && err != ErrUserNotFound {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		if userData.Email != "" && oldUsersEmail != nil {
			if oldUsersEmail.Nickname != nickname {
				return ctx.JSON(http.StatusConflict, ResponseError{Message: "This email is already registered by user: " + oldUsersEmail.Nickname})
			}
		}

		user, err := s.userCache.GetUserByNickname(nickname)
		if user == nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user by nickname: " + nickname})
		}

		if userData.Fullname != "" {
			_, err = s.db.Exec(context.Background(), "UPDATE users SET fullname = $1 "+
				"WHERE nickname = $2", &userData.Fullname, &nickname)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			user.Fullname = userData.Fullname
		}
		if userData.About != "" {
			_, err = s.db.Exec(context.Background(), "UPDATE users SET about = $1 "+
				"WHERE nickname = $2", &userData.About, &nickname)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			user.About = userData.About
		}
		if userData.Email != "" {
			_, err = s.db.Exec(context.Background(), "UPDATE users SET email = $1 "+
				"WHERE nickname = $2", &userData.Email, &nickname)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			user.Email = userData.Email
		}

		if userData.About == "" || userData.Fullname == "" || userData.Email == "" {
			userDataOld, err := s.getUserByNickname(nickname)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			return ctx.JSON(http.StatusOK, userDataOld)
		}

		userData.Nickname = nickname
		s.userCache.Add(user)
		return ctx.JSON(http.StatusOK, &userData)
	}
}
