package internal

import "time"

// Информация о форуме.
type Forum struct {

	// Название форума.
	Title string `json:"title"`

	// Nickname пользователя, который отвечает за форум.
	User string `json:"user"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL), уникальное поле.
	Slug string `json:"slug"`

	// Общее кол-во сообщений в данном форуме.
	Posts uint `json:"posts,omitempty"`

	// Общее кол-во ветвей обсуждения в данном форуме.
	Threads uint `json:"threads,omitempty"`
}

// Сообщение внутри ветки обсуждения на форуме.
type Post struct {

	// Идентификатор данного сообщения.
	Id uint32 `json:"id,omitempty"`

	// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Parent uint `json:"parent,omitempty"`

	// Автор, написавший данное сообщение.
	Author string `json:"author"`

	// Собственно сообщение форума.
	Message string `json:"message"`

	// Истина, если данное сообщение было изменено.
	IsEdited bool `json:"isEdited,omitempty"`

	// Идентификатор форума (slug) данного сообещния.
	Forum string `json:"forum,omitempty"`

	// Идентификатор ветви (id) обсуждения данного сообещния.
	Thread int `json:"thread,omitempty"`

	// Дата создания сообщения на форуме.
	Created time.Time `json:"created,omitempty"`

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
	Id uint `json:"id"`

	// Заголовок ветки обсуждения.
	Title string `json:"title"`

	// Пользователь, создавший данную тему.
	Author string `json:"author"`

	// Форум, в котором расположена данная ветка обсуждения.
	Forum string `json:"forum"`

	// Описание ветки обсуждения.
	Message string `json:"message"`

	// Кол-во голосов непосредственно за данное сообщение форума.
	Votes int `json:"votes"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL). В данной структуре slug опционален и не может быть числом.
	Slug string `json:"slug,omitempty"`

	// Дата создания ветки на форуме.
	Created time.Time `json:"created"`
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
	About string `json:"about,omitempty"`

	// Почтовый адрес пользователя (уникальное поле).
	Email string `json:"email"`

	// Полное имя пользователя.
	Fullname string `json:"fullname"`

	// Имя пользователя (уникальное поле). Данное поле допускает только латиницу, цифры и знак подчеркивания. Сравнение имени регистронезависимо.
	Nickname string `json:"nickname,omitempty"`
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
	Nickname string `json:"nickname"`

	// Отданный голос.
	Voice int `json:"voice"`
}

type ResponseError struct {

	// Текстовое описание ошибки. В процессе проверки API никаких проверок на содерижимое данного описание не делается.
	Message string `json:"message,omitempty"`
}
