//go:generate genopts --file=users.go
package myapp

type User struct {
	Name  string `with:"-"`
	Email string
	Age   int `with:"-"`
}

type SecretUser struct {
	Name  string `with:"-"`
	Email string
	Age   int `with:"-"`
}

type Time struct {
	Nano int64 `with:"-"`
}
