package internal

import "time"

// Информация о форуме.
type Forum struct {

	// Название форума.
	Title string `db:"title" json:"title"`

	// Nickname пользователя, который отвечает за форум.
	User string `db:"author" json:"user"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL), уникальное поле.
	Slug string `db:"slug" json:"slug"`

	// Общее кол-во сообщений в данном форуме.
	Posts uint `db:"posts" json:"posts,omitempty"`

	// Общее кол-во ветвей обсуждения в данном форуме.
	Threads uint `db:"threads" json:"threads,omitempty"`
}

// Сообщение внутри ветки обсуждения на форуме.
type Post struct {

	// Идентификатор данного сообщения.
	Id uint32 `db:"id" json:"id,omitempty"`

	// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Parent uint `db:"parent" json:"parent,omitempty"`

	// Автор, написавший данное сообщение.
	Author string `db:"author" json:"author"`

	// Собственно сообщение форума.
	Message string `db:"message" json:"message"`

	// Истина, если данное сообщение было изменено.
	IsEdited bool `db:"is_edited" json:"isEdited,omitempty"`

	// Идентификатор форума (slug) данного сообещния.
	Forum string `db:"forum" json:"forum,omitempty"`

	// Идентификатор ветви (id) обсуждения данного сообещния.
	ThreadId int `db:"thread_id" json:"thread,omitempty"`

	// Дата создания сообщения на форуме.
	CreatedAt time.Time `db:"created_at" json:"created,omitempty"`

	Path string `json:"-"`
}

// Полная информация о сообщении, включая связанные объекты.
type PostFull struct {
	Post *Post `json:"post,omitempty"`

	Author *User `json:"author,omitempty"`

	Thread *Thread `json:"thread,omitempty"`

	Forum *Forum `json:"forum,omitempty"`
}

// Сообщение для обновления сообщения внутри ветки на форуме. Пустые параметры остаются без изменений.
type PostUpdate struct {

	// Собственно сообщение форума.
	Message string `json:"message,omitempty"`
}

type Status struct {

	// Кол-во пользователей в базе данных.
	User uint `json:"user"`

	// Кол-во разделов в базе данных.
	Forum uint `json:"forum"`

	// Кол-во веток обсуждения в базе данных.
	Thread uint `json:"thread"`

	// Кол-во сообщений в базе данных.
	Post uint `json:"post"`
}

// Ветка обсуждения на форуме.
type Thread struct {

	// Идентификатор ветки обсуждения.
	Id uint `db:"id" json:"id"`

	// Заголовок ветки обсуждения.
	Title string `db:"title" json:"title"`

	// Пользователь, создавший данную тему.
	Author string `db:"author" json:"author"`

	// Форум, в котором расположена данная ветка обсуждения.
	Forum string `db:"forum" json:"forum"`

	// Описание ветки обсуждения.
	Message string `db:"message" json:"message"`

	// Кол-во голосов непосредственно за данное сообщение форума.
	Votes int `db:"votes" json:"votes"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL). В данной структуре slug опционален и не может быть числом.
	Slug string `db:"slug" json:"slug,omitempty"`

	// Дата создания ветки на форуме.
	CreatedAt time.Time `db:"created_at" json:"created"`
}

// Сообщение для обновления ветки обсуждения на форуме. Пустые параметры остаются без изменений.
type ThreadUpdate struct {

	// Заголовок ветки обсуждения.
	Title string `json:"title,omitempty"`

	// Описание ветки обсуждения.
	Message string `json:"message,omitempty"`
}

// Информация о пользователе.
type User struct {
	// Описание пользователя.
	About string `db:"about" json:"about,omitempty"`

	// Почтовый адрес пользователя (уникальное поле).
	Email string `db:"email" json:"email"`

	// Полное имя пользователя.
	Fullname string `db:"fullname" json:"fullname"`

	// Имя пользователя (уникальное поле). Данное поле допускает только латиницу, цифры и знак подчеркивания. Сравнение имени регистронезависимо.
	Nickname string `db:"nickname" json:"nickname,omitempty"`
}

// Информация о пользователе.
type UserUpdate struct {

	// Полное имя пользователя.
	Fullname string `json:"fullname,omitempty"`

	// Описание пользователя.
	About string `json:"about,omitempty"`

	// Почтовый адрес пользователя (уникальное поле).
	Email string `json:"email,omitempty"`
}

// Информация о голосовании пользователя.
type Vote struct {

	// Идентификатор пользователя.
	Nickname string `db:"author" json:"nickname"`

	// Отданный голос.
	Voice int `db:"value" json:"voice"`
}

type ResponseError struct {

	// Текстовое описание ошибки. В процессе проверки API никаких проверок на содерижимое данного описание не делается.
	Message string `json:"message,omitempty"`
}
